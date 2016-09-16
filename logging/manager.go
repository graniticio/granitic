package logging

type ComponentLoggerManager struct {
	created         map[string]*LevelAwareLogger
	initialLevels   map[string]interface{}
	globalThreshold LogLevel
	writers         []LogWriter
	formatter       *LogMessageFormatter
}

func CreateComponentLoggerManager(globalThreshold LogLevel, initalComponentLogLevels map[string]interface{},
	writers []LogWriter, formatter *LogMessageFormatter) *ComponentLoggerManager {

	clm := new(ComponentLoggerManager)
	clm.created = make(map[string]*LevelAwareLogger)
	clm.globalThreshold = globalThreshold
	clm.initialLevels = initalComponentLogLevels

	clm.writers = writers
	clm.formatter = formatter

	return clm
}

func (clm *ComponentLoggerManager) LoggerByName(name string) *LevelAwareLogger {
	return clm.created[name]
}

func (clm *ComponentLoggerManager) CurrentLevels() []*ComponentLevel {

	cls := make([]*ComponentLevel, 0)

	for n, c := range clm.created {

		lev := new(ComponentLevel)
		lev.Level = c.localLogThreshhold
		lev.Name = n

		cls = append(cls, lev)

	}

	return cls
}

func (clm *ComponentLoggerManager) GlobalLevel() LogLevel {
	return clm.globalThreshold
}

func (clm *ComponentLoggerManager) UpdateWritersAndFormatter(writers []LogWriter, formatter *LogMessageFormatter) {
	clm.writers = writers

	for _, v := range clm.created {
		v.UpdateWritersAndFormatter(writers, formatter)
	}
}

func (clm *ComponentLoggerManager) SetGlobalThreshold(globalThreshold LogLevel) {

	clm.globalThreshold = globalThreshold
}

func (clm *ComponentLoggerManager) SetInitialLogLevels(ll map[string]interface{}) {

	clm.initialLevels = ll

	if len(clm.created) > 0 {

		for k, v := range clm.created {

			level := ll[k]

			if level != nil {
				t, _ := LogLevelFromLabel(level.(string))
				v.SetThreshold(t)

			}
		}
	}
}

func (clm *ComponentLoggerManager) CreateLogger(componentId string) Logger {

	if clm.created[componentId] != nil {
		return clm.created[componentId]
	}

	var threshold LogLevel

	threshold = All

	if clm.initialLevels != nil {

		if levelLabel, ok := clm.initialLevels[componentId]; ok {
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

	clm.created[componentId] = l

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

type ComponentLevel struct {
	Name  string
	Level LogLevel
}

type ComponentLevels []*ComponentLevel

func (s ComponentLevels) Len() int      { return len(s) }
func (s ComponentLevels) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

type ByName struct{ ComponentLevels }

func (s ByName) Less(i, j int) bool { return s.ComponentLevels[i].Name < s.ComponentLevels[j].Name }
