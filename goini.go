package goini

import (
	"sync"
	"container/list"
	"path"
	"os"
	"bufio"
	"strings"
	"regexp"
	"fmt"
	"errors"
)

type IniFile struct {
	filePath string
	sections map[string]*list.List
	mutex    sync.RWMutex
	orderedSections []string
}

func NewIniFile(filePathArg string) *IniFile {
	return &IniFile{
		filePath:filePathArg,
		sections:make(map[string]*list.List),
    }
}

//
// Utilities
//

func isSection(section string) bool {
	return strings.HasPrefix(section, "[")
}

// Read parses a specified configuration file and returns a Configuration instance.
func Parse(filePath string) (*IniFile, error) {
	filePath = path.Clean(filePath)
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// New File
	c := NewIniFile(filePath)

	activeSection := c.AddSection("global")

	scanner := bufio.NewScanner(bufio.NewReader(file))
	for scanner.Scan() {
		line := scanner.Text()
		if !(strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";")) && len(line) > 0 {
			if isSection(line) {
				name := strings.Trim(line, " []")
				activeSection = c.AddSection(name)
				continue
			} else {
				activeSection.AddOption(line)
			}
		} else {
			// save comments
			activeSection.AddOption(line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *IniFile) AddSection(name string) *Section {
	section := &Section{name:name, options : make(map[string]string)}
	var lst *list.List
	if lst = c.sections[name]; lst == nil {
		lst = list.New()
		c.sections[name] = lst
		c.orderedSections = append(c.orderedSections, name)
	}

	lst.PushBack(section)
	return section
}

// Save the Configuration to file. Creates a backup (.bak) if file already exists.
func (c *IniFile) Save(filePath string) (err error) {
	c.mutex.Lock()

	err = os.Rename(filePath, filePath+".bak")
	if err != nil {
		if !os.IsNotExist(err) { // fine if the file does not exists
			return err
		}
	}

	f, err := os.Create(filePath)
	if err != nil {
		return err
	}

	defer f.Close()

	w := bufio.NewWriter(f)

	defer w.Flush()

	c.mutex.Unlock()

	s, err := c.Sections()
	if err != nil {
		return err
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	for _, v := range s {
		w.WriteString(v.String())
	}

	return err
}


// FilePath returns the configuration file path.
func (c *IniFile) FilePath() string {
	return c.filePath
}

// StringValue returns the string value for the specified section and option.
func (c *IniFile) StringValue(section, option string) (value string, err error) {
	s, err := c.Section(section)
	if err != nil {
		return
	}
	value = s.ValueOf(option)
	return
}

// Delete deletes the specified sections matched by a regex name and returns the deleted sections.
func (c *IniFile) Delete(regex string) (sections []*Section, err error) {
	sections, err = c.Find(regex)
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if err == nil {
		for _, s := range sections {
			delete(c.sections, s.name)
		}
		// remove also from ordered list
		var matched bool
		for i, name := range c.orderedSections {
			if matched, err = regexp.MatchString(regex, name); matched {
				c.orderedSections = append(c.orderedSections[:i], c.orderedSections[i+1:]...)
			} else {
				if err != nil {
					return nil, err
				}
			}
		}
	}
	return sections, err
}

// Section returns the first section matching the fully qualified section name.
func (c *IniFile) Section(name string) (*Section, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if l, ok := c.sections[name]; ok {
		for e := l.Front(); e != nil; e = e.Next() {
			s := e.Value.(*Section)
			return s, nil
		}
	}
	return nil, errors.New("Unable to find " + name)
}


// Sections returns a slice of Sections matching the fully qualified section name.
func (c *IniFile) Sections(name string) ([]*Section, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	var sections []*Section

	f := func(lst *list.List) {
		for e := lst.Front(); e != nil; e = e.Next() {
			s := e.Value.(*Section)
			sections = append(sections, s)
		}
	}

	if name == "" {
		// Get all sections.
		for _, name := range c.orderedSections {
			if lst, ok := c.sections[name]; ok {
				f(lst)
			}
		}
	} else {
		if lst, ok := c.sections[name]; ok {
			f(lst)
		} else {
			return nil, errors.New("Unable to find " + name)
		}
	}

	return sections, nil
}

// Find returns a slice of Sections matching the regexp against the section name.
func (c *IniFile) Find(regex string) ([]*Section, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	var sections []*Section
	for key, lst := range c.sections {
		if matched, err := regexp.MatchString(regex, key); matched {
			for e := lst.Front(); e != nil; e = e.Next() {
				s := e.Value.(*Section)
				sections = append(sections, s)
			}
		} else {
			if err != nil {
				return nil, err
			}
		}
	}
	return sections, nil
}

// PrintSection prints a text representation of all sections matching the fully qualified section name.
func (c *IniFile) PrintSection(name string) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	sections, err := c.Sections(name)
	if err == nil {
		for _, section := range sections {
			fmt.Print(section)
		}
	} else {
		fmt.Printf("Unable to find section %v\n", err)
	}
}

// String returns the text representation of a parsed configuration file.
func (c *IniFile) String() string {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	var parts []string
	for _, name := range c.orderedSections {
		sections, _ := c.Sections(name)
		for _, section := range sections {
			parts = append(parts, section.String())
		}
	}
	return strings.Join(parts, "")
}
