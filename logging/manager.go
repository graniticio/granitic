package logging

type ComponentLoggerManager struct {
	componentsLogger         map[string]LogThresholdControl
	createdLoggers           map[string]Logger
	InitalComponentLogLevels map[string]interface{}
	globalThreshold          int
}

func CreateComponentLoggerManager(globalThreshold int, initalComponentLogLevels map[string]interface{}) *ComponentLoggerManager {
	loggers := make(map[string]LogThresholdControl)
	manager := new(ComponentLoggerManager)
	manager.componentsLogger = loggers
	manager.createdLoggers = make(map[string]Logger)
	manager.globalThreshold = globalThreshold
	manager.InitalComponentLogLevels = initalComponentLogLevels

	return manager
}

func (clm *ComponentLoggerManager) UpdateGlobalThreshold(globalThreshold int) {
	clm.globalThreshold = globalThreshold

	for _, v := range clm.componentsLogger {
		v.SetGlobalThreshold(globalThreshold)
	}
}

func (clm *ComponentLoggerManager) UpdateLocalThreshold(threshold int) {
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
			threshold = LogLevelFromLabel(levelLabel.(string))
		}

	}

	return clm.CreateLoggerAtLevel(componentId, threshold)
}

func (clm *ComponentLoggerManager) CreateLoggerAtLevel(componentId string, threshold int) Logger {
	logger := new(LevelAwareLogger)
	logger.globalLogThreshold = clm.globalThreshold
	logger.localLogThreshhold = threshold
	logger.loggerName = componentId

	clm.componentsLogger[componentId] = logger
	clm.createdLoggers[componentId] = logger

	return logger
}
