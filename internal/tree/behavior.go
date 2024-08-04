package tree

type Status int

const (
	Running Status = iota
	Failure
	Success
)

type Behavior interface {
	Name() string
	Execute() Status
}

type Task struct {
	Run func() Status
}

func (t *Task) Execute() Status {
	return t.Run()
}
func (t *Task) Name() string {
	return "Task"
}

type Condition struct {
	Check func() Status
}

func (c *Condition) Execute() Status {
	return c.Check()
}
func (c *Condition) Name() string {
	return "Condition"
}

type Sequence struct {
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
func (s *Sequence) Name() string {
	return "Sequence"
}

type Fallback struct {
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
func (f *Fallback) Name() string {
	return "Fallback"
}

type Decorator struct {
	Child    Behavior
	Decorate func(Status) Status
}

func (d *Decorator) Execute() Status {
	return d.Decorate(d.Child.Execute())
}
func (d *Decorator) Name() string {
	return "Decorator"
}

type Parallel struct {
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
func (p *Parallel) Name() string {
	return "Parallel"
}
