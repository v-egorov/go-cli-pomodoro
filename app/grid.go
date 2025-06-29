package app

import (
	"log"

	"github.com/mum4k/termdash/align"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/container/grid"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/terminal/terminalapi"
)

func newGrid(b *buttonsSet, w *widgets, t terminalapi.Terminal) (*container.Container, error) {
	log.Println("newGrid")
	builder := grid.New()
	log.Println("builder created")

	builder.Add(
		grid.RowHeightPerc(40,
			grid.ColWidthPercWithOpts(40,
				[]container.Option{
					container.Border(linestyle.Light),
					container.BorderTitle("Жми Q для выхода"),
				},
				// внутренняя строка
				grid.RowHeightPerc(80,
					grid.Widget(w.donTimer)),
				grid.RowHeightPercWithOpts(20,
					[]container.Option{
						container.AlignHorizontal(align.HorizontalCenter),
					},
					grid.Widget(w.txtTimer,
						container.AlignHorizontal(align.HorizontalCenter),
						container.AlignVertical(align.VerticalMiddle),
						container.PaddingLeftPercent(49),
					),
				),
			),

			grid.ColWidthPerc(60,
				grid.RowHeightPerc(70,
					grid.Widget(w.disType, container.Border(linestyle.Light)),
				),
				grid.RowHeightPerc(30,
					grid.Widget(w.txtInfo, container.Border(linestyle.Light)),
				),
			),
		),
	)

	builder.Add(
		grid.RowHeightPerc(20,
			grid.ColWidthPerc(50, grid.Widget(b.btStart)),
			grid.ColWidthPerc(50, grid.Widget(b.btPause)),
		),
	)

	builder.Add(grid.RowHeightPerc(40))

	gridOpts, err := builder.Build()
	if err != nil {
		return nil, err
	}

	c, err := container.New(t, gridOpts...)
	if err != nil {
		return nil, err
	}
	return c, nil
}
