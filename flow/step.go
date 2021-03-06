package flow

import (
	"io"
	"log"

	"github.com/chrislusf/gleam/instruction"
)

func (fc *FlowContext) NewStep() (step *Step) {
	step = &Step{
		Id:     len(fc.Steps),
		Params: make(map[string]interface{}),
		Meta:   &StepMetadata{IsIdempotent: true},
	}
	fc.Steps = append(fc.Steps, step)
	return
}

func (step *Step) NewTask() (task *Task) {
	task = &Task{Step: step, Id: len(step.Tasks)}
	step.Tasks = append(step.Tasks, task)
	return
}

func (step *Step) SetInstruction(ins instruction.Instruction) {
	step.Name = ins.Name()
	step.Function = ins.Function()
	step.Instruction = ins
}

func (step *Step) RunFunction(task *Task) error {
	var readers []io.Reader
	var writers []io.Writer

	for _, reader := range task.InputChans {
		readers = append(readers, reader.Reader)
	}

	for _, shard := range task.OutputShards {
		writers = append(writers, shard.IncomingChan.Writer)
	}

	defer func() {
		for _, writer := range writers {
			if c, ok := writer.(io.Closer); ok {
				c.Close()
			}
		}
	}()

	task.Stats = &instruction.Stats{}
	err := task.Step.Function(readers, writers, task.Stats)
	if err != nil {
		log.Printf("Failed to run task %s-%d: %v\n", task.Step.Name, task.Id, err)
	}
	return err
}
