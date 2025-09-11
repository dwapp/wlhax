package ui

type Stack struct {
	children []Drawable
}

func NewStack() *Stack {
	return &Stack{}
}

func (stack *Stack) Children() []Drawable {
	return stack.children
}

func (stack *Stack) Push(d Drawable) {
	stack.children = append(stack.children, d)
}

func (stack *Stack) Pop() Drawable {
	if len(stack.children) == 0 {
		return nil
	}

	d := stack.children[len(stack.children)-1]
	stack.children = stack.children[:len(stack.children)-1]
	return d
}

func (stack *Stack) Peek() Drawable {
	if len(stack.children) == 0 {
		return nil
	}

	return stack.children[len(stack.children)-1]
}

func (stack *Stack) Draw(ctx *Context) {
	if child := stack.Peek(); child != nil {
		child.Draw(ctx)
	}
}

func (stack *Stack) Invalidate() {
	Invalidate()
}
