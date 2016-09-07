package logging

type ComponentLoggerManager struct {
	componentsLogger         map[string]LogRuntimeControl
	createdLoggers           map[string]Logger
	InitalComponentLogLevels map[string]interface{}
	globalThreshold          LogLevel
	writers                  []LogWriter
}

func CreateComponentLoggerManager(globalThreshold LogLevel, initalComponentLogLevels map[string]interface{}, writers []LogWriter) *ComponentLoggerManager {
	loggers := make(map[string]LogRuntimeControl)
	clm := new(ComponentLoggerManager)
	clm.componentsLogger = loggers
	clm.createdLoggers = make(map[string]Logger)
	clm.globalThreshold = globalThreshold
	clm.InitalComponentLogLevels = initalComponentLogLevels

	clm.writers = writers

	return clm
}

func (clm *ComponentLoggerManager) UpdateWriters(writers []LogWriter) {
	clm.writers = writers

	for _, v := range clm.componentsLogger {
		v.UpdateWriters(writers)
	}
}

func (clm *ComponentLoggerManager) UpdateGlobalThreshold(globalThreshold LogLevel) {
	clm.globalThreshold = globalThreshold

	for _, v := range clm.componentsLogger {
		v.SetGlobalThreshold(globalThreshold)
	}
}

func (clm *ComponentLoggerManager) UpdateLocalThreshold(threshold LogLevel) {
	clm.globalThreshold = threshold

	for _, v := range clm.componentsLogger {
		v.SetLocalThreshold(threshold)
	}
}

func (clm *ComponentLoggerManager) CreateLogger(componentId string) Logger {

	if clm.createdLoggers[componentId] != nil {
		return clm.createdLoggers[componentId]
	}

	threshold := clm.globalThreshold

	if clm.InitalComponentLogLevels != nil {

		if levelLabel, ok := clm.InitalComponentLogLevels[componentId]; ok {
			t, _ := LogLevelFromLabel(levelLabel.(string))

			threshold = t
		}

	}

	return clm.CreateLoggerAtLevel(componentId, threshold)
}

func (clm *ComponentLoggerManager) CreateLoggerAtLevel(componentId string, threshold LogLevel) Logger {
	l := new(LevelAwareLogger)
	l.globalLogThreshold = clm.globalThreshold
	l.localLogThreshhold = threshold
	l.loggerName = componentId

	clm.componentsLogger[componentId] = l
	clm.createdLoggers[componentId] = l

	l.writers = clm.writers

	return l
}

func (clm *ComponentLoggerManager) PrepareToStop() {

}

func (clm *ComponentLoggerManager) ReadyToStop() (bool, error) {

	for _, w := range clm.writers {
		if w.Busy() {
			return false, nil
		}
	}

	return true, nil
}

func (clm *ComponentLoggerManager) Stop() error {

	for _, w := range clm.writers {
		w.Close()
	}

	return nil
}
