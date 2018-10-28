package outcome

type Sandbox struct {
}

func (s *Sandbox) format() {}

func (s *Sandbox) Validate() error {
	return nil
}

func (s *Sandbox) Reset() {}
