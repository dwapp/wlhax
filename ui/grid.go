package ui

import (
	"math"
)

type Grid struct {
	rows         []GridSpec
	rowLayout    []gridLayout
	columns      []GridSpec
	columnLayout []gridLayout
	cells        []*GridCell
}

const (
	SIZE_EXACT  = iota
	SIZE_WEIGHT = iota
)

// Specifies the layout of a single row or column
type GridSpec struct {
	// One of SIZE_EXACT or SIZE_WEIGHT
	Strategy int
	// Size function
	Size func() int
}

// Used to cache layout of each row/column
type gridLayout struct {
	Offset int
	Size   int
}

type GridCell struct {
	Row     int
	Column  int
	RowSpan int
	ColSpan int
	Content Drawable
}

func NewGrid() *Grid {
	return &Grid{}
}

func (cell *GridCell) At(row, col int) *GridCell {
	cell.Row = row
	cell.Column = col
	return cell
}

func (cell *GridCell) Span(rows, cols int) *GridCell {
	cell.RowSpan = rows
	cell.ColSpan = cols
	return cell
}

func (grid *Grid) Rows(spec []GridSpec) *Grid {
	grid.rows = spec
	return grid
}

func (grid *Grid) Columns(spec []GridSpec) *Grid {
	grid.columns = spec
	return grid
}

func (grid *Grid) AddChild(content Drawable) *GridCell {
	cell := &GridCell{
		RowSpan: 1,
		ColSpan: 1,
		Content: content,
	}
	grid.cells = append(grid.cells, cell)
	return cell
}

func (grid *Grid) Draw(ctx *Context) {
	grid.reflow(ctx)

	for _, cell := range grid.cells {
		if cell.Row >= len(grid.rowLayout) || cell.Column >= len(grid.columnLayout) {
			continue
		}

		rows := grid.rowLayout[cell.Row : cell.Row+cell.RowSpan]
		cols := grid.columnLayout[cell.Column : cell.Column+cell.ColSpan]
		x := cols[0].Offset
		y := rows[0].Offset
		if x < 0 || y < 0 {
			continue
		}

		width := 0
		height := 0
		for _, col := range cols {
			width += col.Size
		}
		for _, row := range rows {
			height += row.Size
		}
		if x+width > ctx.Width() {
			width = ctx.Width() - x
		}
		if y+height > ctx.Height() {
			height = ctx.Height() - y
		}
		if width <= 0 || height <= 0 {
			continue
		}
		subctx := ctx.Subcontext(x, y, width, height)
		if cell.Content != nil {
			cell.Content.Draw(subctx)
		}
	}
}

func (grid *Grid) reflow(ctx *Context) {
	grid.rowLayout = grid.computeLayout(grid.rows, ctx.Height())
	grid.columnLayout = grid.computeLayout(grid.columns, ctx.Width())
}

func (grid *Grid) computeLayout(specs []GridSpec, space int) []gridLayout {
	layout := make([]gridLayout, len(specs))

	// First pass: exact sizes
	remaining := space
	for i, spec := range specs {
		if spec.Strategy == SIZE_EXACT {
			layout[i].Size = spec.Size()
			remaining -= layout[i].Size
		}
	}

	// Second pass: weights
	totalWeight := 0
	for _, spec := range specs {
		if spec.Strategy == SIZE_WEIGHT {
			totalWeight += spec.Size()
		}
	}

	if totalWeight > 0 && remaining > 0 {
		for i, spec := range specs {
			if spec.Strategy == SIZE_WEIGHT {
				layout[i].Size = int(math.Floor(float64(remaining*spec.Size()) / float64(totalWeight)))
			}
		}
	}

	// Set offsets
	offset := 0
	for i := range layout {
		layout[i].Offset = offset
		offset += layout[i].Size
	}

	return layout
}

func (grid *Grid) Invalidate() {
	Invalidate()
}

// Helper function to create constant size functions
func Const(n int) func() int {
	return func() int { return n }
}
