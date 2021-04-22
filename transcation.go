package tiny

import (
	"log"
)

func Transaction(ctx IDataContext, handle func(ctx IDataContext)) {
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
