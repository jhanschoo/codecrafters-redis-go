package resp

type ComSlice struct {
	coms []RESP
}

func NewComSlice() *ComSlice {
	return &ComSlice{coms: nil}
}

func (c *ComSlice) AppendCom(com RESP) {
	c.coms = append(c.coms, com)
}

func (c *ComSlice) RetrieveComs() []RESP {
	coms := c.coms
	c.coms = nil
	return coms
}

func (c *ComSlice) Len() int {
	return len(c.coms)
}

func (c *ComSlice) IsActive() bool {
	return c.coms != nil
}

func (c *ComSlice) Initialize() bool {
	if c.coms == nil {
		c.coms = make([]RESP, 0)
		return true
	}
	return false
}
