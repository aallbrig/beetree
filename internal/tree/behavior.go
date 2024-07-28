package tree

type Status int

const (
	Running Status = iota
	Failure
	Success
)

type Behavior interface {
	Execute() Status
}

type Task struct {
	Name string
	Run  func() Status
}

func (t *Task) Execute() Status {
	return t.Run()
}

type Condition struct {
	Name  string
	Check func() Status
}

func (c *Condition) Execute() Status {
	return c.Check()
}

type Sequence struct {
	Name     string
	Children []Behavior
}

func (s *Sequence) Execute() Status {
	for _, child := range s.Children {
		status := child.Execute()
		if status != Success {
			return status
		}
	}
	return Success
}

type Fallback struct {
	Name     string
	Children []Behavior
}

func (f *Fallback) Execute() Status {
	for _, child := range f.Children {
		status := child.Execute()
		if status == Success {
			return Success
		}
	}
	return Failure
}

type Decorator struct {
	Name     string
	Child    Behavior
	Decorate func(Status) Status
}

func (d *Decorator) Execute() Status {
	return d.Decorate(d.Child.Execute())
}

type Parallel struct {
	Name     string
	Children []Behavior
	Policy   func([]Status) Status
}

func (p *Parallel) Execute() Status {
	results := make([]Status, len(p.Children))
	for i, child := range p.Children {
		results[i] = child.Execute()
	}
	return p.Policy(results)
}
