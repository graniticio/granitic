package logging

type ComponentLoggerManager struct {
	componentsLogger         map[string]RuntimeControllableLog
	createdLoggers           map[string]Logger
	initalComponentLogLevels map[string]interface{}
	globalThreshold          LogLevel
	writers                  []LogWriter
	formatter                *LogMessageFormatter
}

func CreateComponentLoggerManager(globalThreshold LogLevel, initalComponentLogLevels map[string]interface{},
	writers []LogWriter, formatter *LogMessageFormatter) *ComponentLoggerManager {

	loggers := make(map[string]RuntimeControllableLog)
	clm := new(ComponentLoggerManager)
	clm.componentsLogger = loggers
	clm.createdLoggers = make(map[string]Logger)
	clm.globalThreshold = globalThreshold
	clm.initalComponentLogLevels = initalComponentLogLevels

	clm.writers = writers
	clm.formatter = formatter

	return clm
}

func (clm *ComponentLoggerManager) GlobalLevel() LogLevel {
	return clm.globalThreshold
}

func (clm *ComponentLoggerManager) UpdateWritersAndFormatter(writers []LogWriter, formatter *LogMessageFormatter) {
	clm.writers = writers

	for _, v := range clm.componentsLogger {
		v.UpdateWritersAndFormatter(writers, formatter)
	}
}

func (clm *ComponentLoggerManager) SetGlobalThreshold(globalThreshold LogLevel) {

	clm.globalThreshold = globalThreshold
}

func (clm *ComponentLoggerManager) SetInitialLogLevels(ll map[string]interface{}) {

	clm.initalComponentLogLevels = ll

	if len(clm.createdLoggers) > 0 {

		for k, v := range clm.componentsLogger {

			level := ll[k]

			if level != nil {
				t, _ := LogLevelFromLabel(level.(string))
				v.SetThreshold(t)

			}

		}

	}

}

func (clm *ComponentLoggerManager) CreateLogger(componentId string) Logger {

	if clm.createdLoggers[componentId] != nil {
		return clm.createdLoggers[componentId]
	}

	var threshold LogLevel

	threshold = All

	if clm.initalComponentLogLevels != nil {

		if levelLabel, ok := clm.initalComponentLogLevels[componentId]; ok {
			t, _ := LogLevelFromLabel(levelLabel.(string))

			threshold = t
		}

	}

	return clm.CreateLoggerAtLevel(componentId, threshold)
}

func (clm *ComponentLoggerManager) CreateLoggerAtLevel(componentId string, threshold LogLevel) Logger {
	l := new(LevelAwareLogger)
	l.global = clm
	l.localLogThreshhold = threshold
	l.loggerName = componentId

	clm.componentsLogger[componentId] = l
	clm.createdLoggers[componentId] = l

	l.writers = clm.writers
	l.formatter = clm.formatter

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
