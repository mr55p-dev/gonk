package gonk

import "reflect"

type Node struct {
	tag     Tag
	valueOf reflect.Value
}

type Stack struct {
	storage []*Node
}

func (s *Stack) Push(frame *Node) {
	s.storage = append(s.storage, frame)
}

func (s *Stack) Pop() *Node {
	if len(s.storage) == 0 {
		return nil
	}
	end := len(s.storage) - 1
	out := s.storage[end]
	s.storage = s.storage[:end]
	return out
}

func (s *Stack) Size() int {
	return len(s.storage)
}
