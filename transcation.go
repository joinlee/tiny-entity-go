package tiny

import (
	"log"
)

func Transaction[T IDataContext](ctx T, handle func(ctx T)) {
	defer func() {
		if r := recover(); r != nil {
			ctx.RollBack()
			log.Printf("[Tiny] Transaction Error: %v\n", r)
			panic(r)
		}
	}()
	ctx.BeginTranscation()
	handle(ctx)
	ctx.Commit()
}
