package controllers

import "github.com/revel/revel"

type App struct {
	*revel.Controller
}

func (c App) Index() revel.Result {
	c.RenderArgs["walkhub_proxy_url"] = revel.Config.StringDefault("walkthrough.proxy_url", "")
	return c.Render()
}

func (c App) Walkhub() revel.Result {
	return c.Render()
}
