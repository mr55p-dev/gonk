package gonk

import "reflect"

type nodeFrame struct {
	tag     tagData
	valueOf reflect.Value
}

type stack struct {
	storage []*nodeFrame
}

func (s *stack) push(frame ...*nodeFrame) {
	s.storage = append(s.storage, frame...)
}

func (s *stack) pop() *nodeFrame {
	if len(s.storage) == 0 {
		return nil
	}
	end := len(s.storage) - 1
	out := s.storage[end]
	s.storage = s.storage[:end]
	return out
}

func (s *stack) size() int {
	return len(s.storage)
}
