package system

import "fmt"

type DeadLetterProcessor interface {
	Process(actorName string, event interface{})
}

type ConsoleDeadLetterProcessor struct {
}

func (processor *ConsoleDeadLetterProcessor) Process(actorName string, event interface{}) {
	fmt.Printf("Unabled to find actor \"%s\", discard event \"%+v\"", actorName, event)
}

func NewConsoleDeadLetterProcessor() *ConsoleDeadLetterProcessor {
	return &ConsoleDeadLetterProcessor{}
}
