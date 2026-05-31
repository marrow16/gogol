package logic

type Adjacents [8]*Cell

type Cell struct {
	Alive     bool
	Adjacents Adjacents
}

func (c *Cell) AdjacentsAlive() (count int) {
	for _, adj := range c.Adjacents {
		if adj != nil && adj.Alive {
			count++
		}
	}
	return count
}

func (c *Cell) flip() {
	c.Alive = !c.Alive
}
