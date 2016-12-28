package controllers

import (
	"html/template"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/qor/action_bar"
	"github.com/qor/qor-example/app/models"
	"github.com/qor/qor-example/config"
	"github.com/qor/qor-example/config/admin"
	"github.com/qor/seo"
)

func ProductShow(ctx *gin.Context) {
	var (
		product        models.Product
		colorVariation models.ColorVariation
		seoSetting     models.SEOSetting
		codes          = strings.Split(ctx.Param("code"), "_")
		productCode    = codes[0]
		colorCode      string
	)

	if len(codes) > 1 {
		colorCode = codes[1]
	}

	if DB(ctx).Where(&models.Product{Code: productCode}).First(&product).RecordNotFound() {
		http.Redirect(ctx.Writer, ctx.Request, "/", http.StatusFound)
	}

	DB(ctx).Preload("Product").Preload("Color").Preload("SizeVariations.Size").Where(&models.ColorVariation{ProductID: product.ID, ColorCode: colorCode}).First(&colorVariation)
	DB(ctx).First(&seoSetting)

	editButton := admin.ActionBar.RenderEditButton(ctx.Writer, ctx.Request, product)
	config.View.Funcs(funcsMap(ctx)).Execute(
		"product_show",
		gin.H{
			"ActionBarTag":   admin.ActionBar.Render(ctx.Writer, ctx.Request, action_bar.Config{InlineActions: []template.HTML{editButton}}),
			"Product":        product,
			"ColorVariation": colorVariation,
			"SeoTag":         seoSetting.ProductPage.Render(seoSetting, product),
			"MicroProduct": seo.MicroProduct{
				Name:        product.Name,
				Description: product.Description,
				BrandName:   product.Category.Name,
				SKU:         product.Code,
				Price:       float64(product.Price),
				Image:       colorVariation.MainImageURL(),
			}.Render(),
			"CurrentUser":   CurrentUser(ctx),
			"CurrentLocale": CurrentLocale(ctx),
		},
		ctx.Request,
		ctx.Writer,
	)
}

func funcsMap(ctx *gin.Context) template.FuncMap {
	funcMaps := map[string]interface{}{
		"related_products": func(cv models.ColorVariation) []models.Product {
			var products []models.Product
			DB(ctx).Preload("ColorVariations").Limit(4).Find(&products, "id <> ?", cv.ProductID)
			return products
		},
		"other_also_bought": func(cv models.ColorVariation) []models.Product {
			var products []models.Product
			DB(ctx).Preload("ColorVariations").Order("id ASC").Limit(8).Find(&products, "id <> ?", cv.ProductID)
			return products
		},
	}
	for key, value := range I18nFuncMap(ctx) {
		funcMaps[key] = value
	}
	for key, value := range admin.ActionBar.FuncMap(ctx.Writer, ctx.Request) {
		funcMaps[key] = value
	}
	return funcMaps
}
