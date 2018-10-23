package bind

func NewRendererWithDeps(renderEngine chartGoTemplateRenderer, toRenderValuesCaps toRenderValuesCaps) *Renderer {
	return &Renderer{
		renderEngine:       renderEngine,
		toRenderValuesCaps: toRenderValuesCaps,
	}
}
