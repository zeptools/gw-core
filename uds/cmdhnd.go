package uds

import (
	"fmt"
	"io"
)

type CommandHandler interface {
	Command() string // Unique Name
	Desc() string
	Usage() string
	HandleCommand(args []string, w io.Writer) error
}

type CommandGroup struct {
	name         string
	handlerMap   map[string]CommandHandler
	displayOrder []string
}

func NewCommandGroup(name string) *CommandGroup {
	return &CommandGroup{
		name:         name,
		handlerMap:   make(map[string]CommandHandler),
		displayOrder: make([]string, 0),
	}
}

func (g *CommandGroup) Add(handler CommandHandler) error {
	cmd := handler.Command()
	_, exists := g.handlerMap[cmd]
	if exists {
		return fmt.Errorf("command %q already exists", cmd)
	}
	g.handlerMap[cmd] = handler
	g.displayOrder = append(g.displayOrder, cmd)
	return nil
}

func (g *CommandGroup) AddMany(handlers []CommandHandler) error {
	for _, handler := range handlers {
		if err := g.Add(handler); err != nil {
			return err
		}
	}
	return nil
}

type CommandStore struct {
	handlerMap        map[string]CommandHandler
	groupMap          map[string]*CommandGroup
	groupDisplayOrder []string
}

func NewCommandStore() *CommandStore {
	return &CommandStore{
		handlerMap:        make(map[string]CommandHandler),
		groupMap:          make(map[string]*CommandGroup),
		groupDisplayOrder: make([]string, 0),
	}
}

func (s *CommandStore) Add(cmdGroup *CommandGroup) error {
	_, exists := s.groupMap[cmdGroup.name]
	if exists {
		return fmt.Errorf("command group %q already exists", cmdGroup.name)
	}
	s.groupMap[cmdGroup.name] = cmdGroup
	s.groupDisplayOrder = append(s.groupDisplayOrder, cmdGroup.name)
	for cmd, handler := range cmdGroup.handlerMap {
		_, exists = s.handlerMap[cmd]
		if exists {
			return fmt.Errorf("command %q already exists", cmd)
		}
		s.handlerMap[cmd] = handler
	}
	return nil
}

func (s *CommandStore) AddMany(cmdGroups []*CommandGroup) error {
	for _, cmdGroup := range cmdGroups {
		if err := s.Add(cmdGroup); err != nil {
			return err
		}
	}
	return nil
}

func (s *CommandStore) GetHandler(cmd string) (CommandHandler, bool) {
	handler, ok := s.handlerMap[cmd]
	return handler, ok
}

func (s *CommandStore) PrintHelp(w io.Writer) {
	_, _ = fmt.Fprintln(w)
	for _, grpName := range s.groupDisplayOrder {
		cmdGrp, ok := s.groupMap[grpName]
		if !ok {
			continue
		}
		_, _ = fmt.Fprintf(w, "---- %s ----\n", grpName)
		for _, cmd := range cmdGrp.displayOrder {
			cmdHandler, ok := cmdGrp.handlerMap[cmd]
			if !ok {
				continue
			}
			_, _ = fmt.Fprintf(w, "%-36s %s\n", cmd, cmdHandler.Desc())
		}
		_, _ = fmt.Fprintln(w)
	}
	_, _ = fmt.Fprintln(w)
}

type CmdHnd struct {
	Desc  string
	Usage string
	Fn    func(args []string, w io.Writer) error
}
