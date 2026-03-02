# usecase

> Use case implementation - imports ports/out, returns ports/in executor

## app/application/usecase/get_template.go

```go
package usecase

import (
	"context"

	"archetype/app/application/ports/in"
	"archetype/app/application/ports/out"

	"github.com/Ignaciojeria/ioc"
)

var _ = ioc.Register(NewGetTemplateUseCase)

type getTemplateUseCase struct {
	repo out.TemplateRepository
}

func NewGetTemplateUseCase(repo out.TemplateRepository) (in.GetTemplateExecutor, error) {
	return &getTemplateUseCase{repo: repo}, nil
}

func (uc *getTemplateUseCase) Execute(ctx context.Context, id string) (in.GetTemplateOutput, error) {
	t, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return in.GetTemplateOutput{}, err
	}
	return in.GetTemplateOutput{ID: t.ID}, nil
}
```
