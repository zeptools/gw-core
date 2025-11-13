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

func NewCommandGroup(name string, handlers []CommandHandler) *CommandGroup {
	g := &CommandGroup{
		name:         name,
		handlerMap:   make(map[string]CommandHandler),
		displayOrder: make([]string, 0),
	}
	g.AddMany(handlers)
	return g
}

func (g *CommandGroup) Add(handler CommandHandler) {
	cmd := handler.Command()
	if _, exists := g.handlerMap[cmd]; exists {
		return
	}
	g.handlerMap[cmd] = handler
	g.displayOrder = append(g.displayOrder, cmd)
}

func (g *CommandGroup) AddMany(handlers []CommandHandler) {
	for _, handler := range handlers {
		g.Add(handler)
	}
}

type CommandStore struct {
	handlerMap        map[string]CommandHandler
	groupMap          map[string]*CommandGroup
	groupDisplayOrder []string
}

func NewCommandStore(cmdGroups []*CommandGroup) *CommandStore {
	s := &CommandStore{
		handlerMap:        make(map[string]CommandHandler),
		groupMap:          make(map[string]*CommandGroup),
		groupDisplayOrder: make([]string, 0),
	}
	s.AddMany(cmdGroups)
	return s
}

func (s *CommandStore) Add(cmdGroup *CommandGroup) {
	_, exists := s.groupMap[cmdGroup.name]
	if exists {
		return
	}
	s.groupDisplayOrder = append(s.groupDisplayOrder, cmdGroup.name)
	s.groupMap[cmdGroup.name] = cmdGroup

	for cmd, handler := range cmdGroup.handlerMap {
		if _, exists := s.handlerMap[cmd]; exists {
			continue
		}
		s.handlerMap[cmd] = handler
	}
}

func (s *CommandStore) AddMany(cmdGroups []*CommandGroup) {
	for _, cmdGroup := range cmdGroups {
		s.Add(cmdGroup)
	}
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
