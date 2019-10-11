// Copyright 2016-2019 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package logging

// CreateComponentLoggerManager creates a new ComponentLoggerManager with a global level and default values
// for named components.
func CreateComponentLoggerManager(globalThreshold LogLevel, initalComponentLogLevels map[string]interface{},
	writers []LogWriter, formatter *LogMessageFormatter) *ComponentLoggerManager {

	clm := new(ComponentLoggerManager)
	clm.created = make(map[string]*GraniticLogger)
	clm.globalThreshold = globalThreshold
	clm.initialLevels = initalComponentLogLevels

	clm.writers = writers
	clm.formatter = formatter

	return clm
}

// ComponentLoggerManager creates new Logger instances for a particular scope (e.g. framework or application).
type ComponentLoggerManager struct {
	created         map[string]*GraniticLogger
	initialLevels   map[string]interface{}
	globalThreshold LogLevel
	writers         []LogWriter
	formatter       *LogMessageFormatter
	disabled        bool
	nullLogger      Logger
}

// LoggerByName finds a previously created Logger by the name it was given when it was created. Returns nil if no Logger
// by that name exists.
func (clm *ComponentLoggerManager) LoggerByName(name string) *GraniticLogger {
	return clm.created[name]
}

// CurrentLevels returns the current local log level for all Loggers managed by this component.
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

// Disable prevents this manager from creating new loggers - a NullLogger will be returned instead
func (clm *ComponentLoggerManager) Disable() {
	clm.disabled = true
	clm.nullLogger = new(NullLogger)

}

// IsDisabled checks to see if this manager has been disabled
func (clm *ComponentLoggerManager) IsDisabled() bool {
	return clm.disabled
}

// GlobalLevel returns the global log level for the scope (application, framework) that this ComponentLoggerManager is responsible for.
func (clm *ComponentLoggerManager) GlobalLevel() LogLevel {
	return clm.globalThreshold
}

// UpdateWritersAndFormatter updates the writers and formatters of all Loggers managed by this ComponentLoggerManager.
func (clm *ComponentLoggerManager) UpdateWritersAndFormatter(writers []LogWriter, formatter *LogMessageFormatter) {
	clm.writers = writers

	for _, v := range clm.created {
		v.UpdateWritersAndFormatter(writers, formatter)
	}
}

// SetGlobalThreshold sets the global log level for the scope (application, framework) that this ComponentLoggerManager is responsible for.
func (clm *ComponentLoggerManager) SetGlobalThreshold(globalThreshold LogLevel) {

	clm.globalThreshold = globalThreshold
}

// SetInitialLogLevels provide a map of component names to log levels. If a Logger is subsequently created for a component named in the map,
// the log level in the map will be used to set its local log threshold.
// Previously created loggers will be updated
func (clm *ComponentLoggerManager) SetInitialLogLevels(ll map[string]interface{}) {

	clm.initialLevels = ll

	if len(clm.created) > 0 {

		for k, v := range clm.created {

			level := ll[k]

			if level != nil {
				t, _ := LogLevelFromLabel(level.(string))
				v.SetLocalThreshold(t)

			}
		}
	}
}

// CreateLogger creates a new Logger for the supplied component name
func (clm *ComponentLoggerManager) CreateLogger(componentID string) Logger {

	if clm.disabled {
		return clm.nullLogger
	}

	if clm.created[componentID] != nil {
		return clm.created[componentID]
	}

	var threshold LogLevel

	threshold = All

	if clm.initialLevels != nil {

		if levelLabel, ok := clm.initialLevels[componentID]; ok {
			t, _ := LogLevelFromLabel(levelLabel.(string))

			threshold = t
		}

	}

	return clm.CreateLoggerAtLevel(componentID, threshold)
}

// CreateLoggerAtLevel creates a new Logger for the supplied component name with the local log threshold set to the supplied level.
func (clm *ComponentLoggerManager) CreateLoggerAtLevel(componentID string, threshold LogLevel) Logger {

	if clm.disabled {
		return clm.nullLogger
	}

	l := new(GraniticLogger)
	l.global = clm
	l.localLogThreshhold = threshold
	l.loggerName = componentID

	clm.created[componentID] = l

	l.writers = clm.writers
	l.formatter = clm.formatter

	return l
}

// PrepareToStop does nothing
func (clm *ComponentLoggerManager) PrepareToStop() {

}

// ReadyToStop returns false if any of the LogWriters attached to this component are actively writing.
func (clm *ComponentLoggerManager) ReadyToStop() (bool, error) {

	for _, w := range clm.writers {
		if w.Busy() {
			return false, nil
		}
	}

	return true, nil
}

// Stop closes all LogWriters attached to this component.
func (clm *ComponentLoggerManager) Stop() error {

	for _, w := range clm.writers {
		w.Close()
	}

	return nil
}

// ComponentLevel pairs a component name and its loglevel for sorting and presentation through RuntimeCtl
type ComponentLevel struct {
	Name  string
	Level LogLevel
}

// ComponentLevels allows a slice of ComponentLevel structs to be sorted
type ComponentLevels []*ComponentLevel

// Len supports sorting
func (s ComponentLevels) Len() int { return len(s) }

// Swap supports sorting
func (s ComponentLevels) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

// ByName allows ComponentLevels to be sorted by name
type ByName struct{ ComponentLevels }

// Less supports sorting by name
func (s ByName) Less(i, j int) bool { return s.ComponentLevels[i].Name < s.ComponentLevels[j].Name }
