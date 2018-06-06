package goini

import (
	"strings"
	"sync"
)


type Section struct {
	name string
	options map[string]string
	mutex sync.RWMutex
	orderedOptions []string
}

// Name returns the name of the section
func (s *Section) Name() string {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.name
}

// Exists returns true if the option exists
func (s *Section) Exists(option string) (ok bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	_, ok = s.options[option]
	return
}

// ValueOf returns the value of specified option.
func (s *Section) ValueOf(option string) string {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.options[option]
}

// SetValueFor sets the value for the specified option and returns the old value.
func (s *Section) SetValueFor(option string, value string) string {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	var oldValue string
	oldValue, s.options[option] = s.options[option], value

	return oldValue
}

// Add adds a new option to the section. Adding and existing option will overwrite the old one.
// The old value is returned
func (s *Section) Add(option string, value string) (oldValue string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	var ok bool
	if oldValue, ok = s.options[option]; !ok {
		s.orderedOptions = append(s.orderedOptions, option)
	}
	s.options[option] = value

	return oldValue
}

// Delete removes the specified option from the section and returns the deleted option's value.
func (s *Section) Delete(option string) (value string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	value = s.options[option]
	delete(s.options, option)
	for i, opt := range s.orderedOptions {
		if opt == option {
			s.orderedOptions = append(s.orderedOptions[:i], s.orderedOptions[i+1:]...)
		}
	}
	return value
}

// Options returns a map of options for the section.
func (s *Section) Options() map[string]string {
	return s.options
}

// OptionNames returns a slice of option names in the same order as they were parsed.
func (s *Section) OptionNames() []string {
	return s.orderedOptions
}

// String returns the text representation of a section with its options.
func (s *Section) String() string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	var parts []string
	sName := "[" + s.name + "]\n"
	if s.name == "global" {
		sName = ""
	}
	parts = append(parts, sName)

	for _, opt := range s.orderedOptions {
		value := s.options[opt]
		if value != "" {
			parts = append(parts, opt, "=", value, "\n")
		} else {
			parts = append(parts, opt, "\n")
		}
	}

	return strings.Join(parts, "")
}

func parseOption(option string) (opt, value string) {

	split := func(i int, delim string) (opt, value string) {
		opt = strings.Trim(option[:i], " ")
		value = strings.Trim(option[i+1:], " ")
		return
	}

	if i := strings.Index(option, "="); i != -1 {
		opt, value = split(i, "=")
	} else if i := strings.Index(option, ":"); i != -1 {
		opt, value = split(i, ":")
	} else {
		opt = option
	}
	return
}

//Section object
func (s *Section) AddOption(option string){
	var opt, value string
	if opt, value = parseOption(option); value != "" {
		s.options[opt] = value
	}
}
