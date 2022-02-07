package logger

//---------------------------------------------------------------------------------------------------
// Building a json logger. Takes the content from the methods and converts to JSON and prints to
// the screen
//---------------------------------------------------------------------------------------------------

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

//ensure that logger implements the Logger interface
var (
	_ Logger = &logger{}
	_ Owner  = &logger{}
	_ Manage = &logger{}
)

//Logger interfacce
type Logger interface {
	Debug(content string)
	Info(content string)
	Warn(content string)
	Error(content error)
	FormatError(format string, errs ...interface{})
	Fatal(content string)
	DebugService(serviceName, content string)
	InfoService(serviceName, content string)
	WarnService(serviceName, content string)
	ErrorService(serviceName string, content error)
	FormatErrorService(serviceName, format string, errs ...interface{})
	FatalService(serviceName, content string)
	IsDebugEnabled(serviceName ...string) bool
	EnableDebug(serviceName ...string)
	DisableDebug(serviceName ...string)
	UpdateDebugMap(serviceName string, status bool)
	SetSystemDebugStatus(bool)
	GetSystemDebugStatus() bool
	CheckDebugMap(serviceName string) bool
}

type ServiceDebug struct {
	debugMode map[string]bool
	mu        *sync.RWMutex
}

// Owner interface
type Owner interface {
	Configure(commonName string, envs map[string]string)

	Close()
}

// Manage interface
type Manage interface {
	//Start
	Start() (err error)
	//Stop
	Stop() (err error)
}

//Logger - Defines the logger object
type logger struct {
	sync.WaitGroup
	sync.RWMutex //mutex for threadsafe operations
	started      bool
	commonName   string //used for functions called by common components
	stopper      chan struct{}
	systemDebug  bool
	debugModeMap ServiceDebug
	debug        chan bool
}

// NewLogger returns interfacce
func NewLogger() interface {
	Logger
	Owner
	Manage
} {
	debug := make(chan bool)
	return &logger{
		debug: debug,
	}
}

func (l *logger) Configure(commonName string, envs map[string]string) {
	l.Lock()
	defer l.Unlock()

	//get configuration
	ConfigureFromEnv(envs)
	//set common component name
	l.commonName = commonName
	if runtime.GOOS == "windows" {
		path := "log/"
		if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
			err := os.Mkdir(path, os.ModePerm)
			if err != nil {
				log.Println(err)
			}
		}
		InitLogging("log/" + commonName + ".log")
	} else {
		InitLogging(commonName + ".log")
	}
	//create the debugger map
	l.debugModeMap = ServiceDebug{
		debugMode: make(map[string]bool),
		mu:        &sync.RWMutex{},
	}

}

func (l *logger) Close() {
	l.Lock()
	defer l.Unlock()

	//close internal pointers
	close(l.debug)
	//set internal pointers to nil
	l.stopper, l.debug = nil, nil
}

//Start
func (l *logger) Start() (err error) {
	l.Lock()
	defer l.Unlock()

	if l.started {
		return
	}

	l.stopper = make(chan struct{})
	//launch debug
	l.LaunchDebug()
	//set started to true
	l.started = true

	return
}

//Stop
func (l *logger) Stop() (err error) {
	l.Lock()
	defer l.Unlock()

	if !l.started {
		return
	}
	l.debugModeMap = ServiceDebug{
		debugMode: make(map[string]bool),
		mu:        &sync.RWMutex{}}
	//close stopper
	close(l.stopper)
	//wait for goRoutines to return
	l.Wait()
	//set started to false
	l.started = false

	return
}

//GetDebugTime - This method retrieves the time for which debug will be run
func (l *logger) GetDebugTime() time.Duration {
	l.RLock()
	defer l.RUnlock()

	return ConfigDebugTimer
}

func (l *logger) UpdateDebugMap(serviceName string, status bool) {
	l.debugModeMap.mu.Lock()
	l.debugModeMap.debugMode[serviceName] = status
	l.debugModeMap.mu.Unlock()
}

//IsDebugEnabled - Returns the debug mode
func (l *logger) IsDebugEnabled(serviceName ...string) (debugmode bool) {
	l.RLock()
	defer l.RUnlock()

	if len(serviceName) != 0 {
		l.debugModeMap.mu.Lock()
		if mode, ok := l.debugModeMap.debugMode[serviceName[0]]; ok {
			debugmode = mode
		}
		l.debugModeMap.mu.Unlock()
	} else {
		debugmode = l.systemDebug
	}

	return
}

//EnableDebug - Enable Debug if not set
func (l *logger) EnableDebug(serviceName ...string) {
	// l.Lock()
	// defer l.Unlock()
	if len(serviceName) != 0 {
		l.debugModeMap.mu.Lock()
		if _, ok := l.debugModeMap.debugMode[serviceName[0]]; ok {
			l.debugModeMap.debugMode[serviceName[0]] = true
		}
		l.debugModeMap.mu.Unlock()
		l.debug <- true
	} else {
		//if service name not specified enable everything
		l.SetSystemDebugStatus(true)
	}
}

//DisableDebug - Disable Debug if set
func (l *logger) DisableDebug(serviceName ...string) {
	// l.RLock()
	// defer l.RUnlock()
	if len(serviceName) != 0 {
		l.debugModeMap.mu.Lock()
		if _, ok := l.debugModeMap.debugMode[serviceName[0]]; ok {
			l.debugModeMap.debugMode[serviceName[0]] = false
		}
		l.debugModeMap.mu.Unlock()
		l.debug <- false
	} else {
		//if service name not specified disable everything
		l.SetSystemDebugStatus(false)
	}
}

//SetSystemDebugStatus - function to set if the call is a system call or service call
func (l *logger) SetSystemDebugStatus(status bool) {
	//set all individual service status
	l.debugModeMap.mu.Lock()
	for serviceName := range l.debugModeMap.debugMode {
		l.debugModeMap.debugMode[serviceName] = status
	}
	l.debugModeMap.mu.Unlock()
	//set overall status
	l.systemDebug = status
	l.debug <- status
}

func (l *logger) GetSystemDebugStatus() bool {
	return l.systemDebug
}

func (l *logger) CheckDebugMap(serviceName string) (found bool) {
	_, found = l.debugModeMap.debugMode[serviceName]

	return
}

func (l *logger) LaunchDebug() {
	started := make(chan struct{})
	l.Add(1)
	go l.goDebug(started)
	<-started
}

//goDebug - Creates a routine to manage debug time
func (l *logger) goDebug(started chan struct{}) {
	defer l.Done()

	expire := time.NewTicker(ConfigDebugTimer)
	expire.Stop()
	defer expire.Stop()
	close(started)

	for {
		select {
		case <-l.stopper:
			return

		case <-expire.C:
			expire.Stop()
			l.Info("Debug Timer Expired")
			//time expired, reset all services debug mode and overall
			l.debugModeMap.mu.Lock()
			for serviceName := range l.debugModeMap.debugMode {
				l.debugModeMap.debugMode[serviceName] = false
			}
			l.debugModeMap.mu.Unlock()
			l.systemDebug = false

		case debug := <-l.debug:
			//setting to debug will reset the timer
			if debug {
				expire.Stop()
				expire = time.NewTicker(ConfigDebugTimer)
				l.Info("Debug Mode Enabled")
			} else {
				//only stop the timer if no other service is in debug mode
				otherDebug := false
				l.debugModeMap.mu.Lock()
				for _, mode := range l.debugModeMap.debugMode {
					if mode {
						otherDebug = true
						break
					}
				}
				l.debugModeMap.mu.Unlock()
				if !otherDebug {
					expire.Stop()
					l.Info("Debug Mode Disabled")
				}
			}
		}
	}
}

//Performs the actual logging operation
func (l *logger) log(serviceName, content string, severity string) {
	//Make the entry
	logentry := LogEntryDocker{Level: severity,
		Timestamp: time.Now().Unix(),
		Name:      serviceName,
		Content:   content,
	}
	entry := LogEntry{
		Name:    serviceName,
		Content: content,
	}
	severity = strings.ToUpper(severity)
	//Marshal the struct
	logjson, _ := json.Marshal(entry)
	dockerjson, _ := json.Marshal(logentry)
	//Print the JSON
	fmt.Println(string(dockerjson))
	if severity == "DEBUG" {
		ZapLogger.Debug(string(logjson))
	} else if severity == "INFO" {
		ZapLogger.Info(string(logjson))
	} else if severity == "WARN" {
		ZapLogger.Warn(string(logjson))
	} else {
		ZapLogger.Error(string(logjson))

	}

}

//Debug
func (l *logger) Debug(content string) {
	if l.systemDebug {
		l.log(l.commonName, content, DEBUG)
	}
}

//Info
func (l *logger) Info(content string) {
	l.log(l.commonName, content, INFO)
}

//Warn
func (l *logger) Warn(content string) {
	l.log(l.commonName, content, WARN)
}

//Error
func (l *logger) Error(content error) {
	l.log(l.commonName, content.Error(), ERROR)
}

//FormatError
func (l *logger) FormatError(format string, errs ...interface{}) {
	l.log(l.commonName, fmt.Errorf(format, errs...).Error(), ERROR)
}

//Fatal
func (l *logger) Fatal(content string) {
	l.log(l.commonName, content, FATAL)
}

//DebugService
func (l *logger) DebugService(serviceName, content string) {
	if l.systemDebug {
		l.log(serviceName, content, DEBUG)
	} else {
		l.debugModeMap.mu.Lock()
		if mode, ok := l.debugModeMap.debugMode[serviceName]; ok {
			if mode {
				l.log(serviceName, content, DEBUG)
			}
		}
		l.debugModeMap.mu.Unlock()
	}
}

//InfoService
func (l *logger) InfoService(serviceName, content string) {
	l.log(serviceName, content, INFO)
}

//WarnService
func (l *logger) WarnService(serviceName, content string) {
	l.log(serviceName, content, WARN)
}

//ErrorService
func (l *logger) ErrorService(serviceName string, content error) {
	l.log(serviceName, content.Error(), ERROR)
}

//FormatErrorService
func (l *logger) FormatErrorService(serviceName, format string, errs ...interface{}) {
	l.log(serviceName, fmt.Errorf(format, errs...).Error(), ERROR)
}

//FatalService
func (l *logger) FatalService(serviceName, content string) {
	l.log(serviceName, content, FATAL)
}
