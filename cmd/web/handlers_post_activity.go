package main

import (
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/davidkuda/bellevue/internal/models"
)

var (
	FieldError = errors.New("FieldError")
)

type productForm struct {
	ID          int
	UserID      int
	Date        time.Time
	Products    []parsedProduct
	Comment     string
	FieldErrors map[string]string
}

type parsedProduct struct {
	Code          string
	PriceCategory string
	Price         int
	Quantity      int
	AmountCHF     int // for snacks
}

// POST /activity
func (app *application) bellevueActivityPost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Printf("Failed parsing form: %v", err)
		app.renderClientError(w, r, http.StatusBadRequest)
		return
	}

	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		log.Println("post /bellevue-activity: could not get userID from request.Context")
		app.renderClientError(w, r, http.StatusUnauthorized)
		return
	}

	formNew := app.parseProductForm(r)
	formNew.UserID = userID

	// TODO: if ValidationErrors, return form with errors
	if len(formNew.FieldErrors) > 0 {
		t := app.newTemplateData(r)
		app.render(w, r, http.StatusUnprocessableEntity, "activities.new.tmpl.html", &t)
		return
	}

	// TODO: deal with taxes
	var consumptions models.Consumptions
	for _, p := range formNew.Products {
		var pricecat string
		if p.PriceCategory != "" {
			pricecat = "/" + p.PriceCategory
		}
		price := p.Price
		if price == 0 {
			price = p.AmountCHF
		}
		consumption := models.Consumption{
			UserID:    userID,
			ProductID: app.productIDMap[p.Code + pricecat],
			TaxID:     0,
			Price:     price,
			TaxPrice:  0,
			Date:      formNew.Date,
		}
		consumptions = append(consumptions, consumption)
	}

	err = app.models.Consumptions.InsertMany(userID, formNew.Date, consumptions)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	if formNew.Comment != "" {
		// TODO: insert comment
	}

	form := bellevueActivityForm{}

	b := form.toModel()

	b.UserID = userID

	err = app.models.BellevueActivities.Insert(b)
	if err != nil {
		log.Printf("app.bellevueActivities.Insert(): %v\n", err)
		app.serverError(w, r, err)
		return
	}

	// TODO: send some notification (Toast) to the UI (successfully submitted)
	http.Redirect(w, r, "/bellevue-activities", http.StatusSeeOther)
	return
}

func (app *application) parseProductForm(r *http.Request) productForm {
	form := productForm{}
	form.FieldErrors = map[string]string{}

	dateStr := r.PostForm.Get("date")
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		form.FieldErrors["date"] = "invalid date input"
	}
	form.Date = date

	// NOTE: an alternative could be iterating over the key-value-pairs of the r.Form
	// for key, values := range r.Form {
	// 	for _, v := range values {
	// 		fmt.Printf("key=%s value=%s\n", key, v)
	// 	}
	// }

	for _, productFormSpec := range app.productFormConfig.Specs {
		var pp parsedProduct
		pp.Code = productFormSpec.Code
		if productFormSpec.HasCategories {
			quantityStr := r.FormValue("activities[" + productFormSpec.Code + "][quantity]")
			quantityInt, err := strconv.Atoi(quantityStr)
			if err != nil {
				form.FieldErrors[productFormSpec.Code+"-Atoi"] = "input is not a number"
				continue
			}
			if quantityInt < 0 {
				form.FieldErrors[productFormSpec.Code+"-Atoi"] = "input is a negative number"
				continue
			}
			if quantityInt == 0 {
				continue
			}
			pp.Quantity = quantityInt

			pricecatFormField := fmt.Sprintf("activities[%s][price_category]", productFormSpec.Code)
			pricecat := r.FormValue(pricecatFormField)
			if ok := app.priceCategoryMap[pricecat]; ok != true {
				form.FieldErrors[productFormSpec.Code+"-price-category"] = "invalid price category"
			}
			pp.PriceCategory = pricecat
			pp.Price = app.productFormConfig.Prices[pp.Code+"/"+pricecat]
		}

		if productFormSpec.IsCustomAmount {
			priceStr := r.FormValue("activities[" + productFormSpec.Code + "][amount_chf]")
			// default input is 0, ignore if 0
			if priceStr == "0" {
				continue
			}

			var priceInt int
			priceFloat, err := strconv.ParseFloat(priceStr, 64)
			if err != nil {
				form.FieldErrors[productFormSpec.Code+"-ParseFloat"] = "invalid custom amount CHF"
				continue
			}
			priceInt = int(math.Round(priceFloat * 100))

			if priceInt < 0 {
				form.FieldErrors[productFormSpec.Code] = "input is a negative number"
				continue
			}

			pp.AmountCHF = priceInt
		}
		form.Products = append(form.Products, pp)
	}

	form.Comment = r.PostForm.Get("comment")

	return form
}

type bellevueActivityForm struct {
	ID                     int
	UserID                 int
	Date                   time.Time
	Breakfasts             int
	BreakfastPriceCategory string
	Lunches                int
	LunchPriceCategory     string
	Dinners                int
	DinnerPriceCategory    string
	Coffees                int
	CoffeePriceCategory    string
	Saunas                 int
	SaunaPriceCategory     string
	Lectures               int
	LecturePriceCategory   string
	SnacksCHF              int
	Comment                string
	FieldErrors            map[string]string
}

func (form *bellevueActivityForm) parseFormFromRequest(r *http.Request) error {
	f := r.PostForm

	dateStr := f.Get("date")
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return fmt.Errorf("form.Date: stconv.Atoi: someone wants to write non-integers: value: %v, err: %v", f.Get("bellevue-activity-breakfasts"), err)
	}

	errorTemplate := "form.%v: stconv.Atoi: someone wants to write non-integers: value: %v, err: %v"

	// although the form has client side validation of integers by using
	// <input type="number">, a malicious actor could still place a POST
	// request not via the web form.
	breakfasts, err := strconv.Atoi(f.Get("breakfasts"))
	if err != nil {
		return fmt.Errorf(errorTemplate, "Breakfasts", f.Get("breakfasts"), err)
	}

	lunches, err := strconv.Atoi(f.Get("lunches"))
	if err != nil {
		return fmt.Errorf(errorTemplate, "Lunches", f.Get("lunches"), err)
	}

	dinners, err := strconv.Atoi(f.Get("dinners"))
	if err != nil {
		return fmt.Errorf(errorTemplate, "Dinners", f.Get("dinners"), err)
	}

	coffees, err := strconv.Atoi(f.Get("coffees"))
	if err != nil {
		return fmt.Errorf(errorTemplate, "Coffees", f.Get("coffees"), err)
	}

	saunas, err := strconv.Atoi(f.Get("saunas"))
	if err != nil {
		return fmt.Errorf(errorTemplate, "Saunas", f.Get("saunas"), err)
	}

	lectures, err := strconv.Atoi(f.Get("lectures"))
	if err != nil {
		return fmt.Errorf(errorTemplate, "Lectures", f.Get("lectures"), err)
	}

	var snacksCHF int
	snackCHFString := f.Get("snacks")
	if len(snackCHFString) > 0 {
		priceFloat, err := strconv.ParseFloat(snackCHFString, 64)
		if err != nil {
			return fmt.Errorf("failed parsing string \"%s\" to float:", snackCHFString)
		}
		snacksCHF = int(math.Round(priceFloat * 100))
	}

	form.Date = date
	form.Breakfasts = breakfasts
	form.BreakfastPriceCategory = f.Get("price-category-breakfasts")
	form.Lunches = lunches
	form.LunchPriceCategory = f.Get("price-category-lunches")
	form.Dinners = dinners
	form.DinnerPriceCategory = f.Get("price-category-dinners")
	form.Coffees = coffees
	form.CoffeePriceCategory = f.Get("price-category-coffees")
	form.Saunas = saunas
	form.SaunaPriceCategory = f.Get("price-category-saunas")
	form.Lectures = lectures
	form.LecturePriceCategory = f.Get("price-category-lectures")
	form.SnacksCHF = snacksCHF
	form.Comment = f.Get("bellevue-activity-comment")
	form.FieldErrors = map[string]string{}

	if form.hasNegativeNumbers() {
		form.FieldErrors["negatives"] = "you may not send negative numbers."
	}

	if form.hasOnlyZeroes() {
		form.FieldErrors["zeroes"] = "you have all 0 and therefore not any activity to upload."
	}

	// TODO: implement form validation for price category
	for _, check := range []string{
		form.BreakfastPriceCategory,
		form.LunchPriceCategory,
		form.DinnerPriceCategory,
		form.CoffeePriceCategory,
		form.SaunaPriceCategory,
		form.LecturePriceCategory,
	} {
		fmt.Println(check)
		if check != "reduced" && check != "regular" && check != "surplus" {
			form.FieldErrors["pricecat"] = "you have an invalid price category somewhere"
		}
	}

	if len(form.FieldErrors) > 0 {
		return FieldError
	}

	return nil
}

func (b *bellevueActivityForm) hasNegativeNumbers() bool {
	if b.Breakfasts < 0 {
		return true
	}
	if b.Lunches < 0 {
		return true
	}
	if b.Dinners < 0 {
		return true
	}
	if b.Coffees < 0 {
		return true
	}
	if b.Saunas < 0 {
		return true
	}
	if b.Lectures < 0 {
		return true
	}
	if b.SnacksCHF < 0 {
		return true
	}
	return false
}

func (b *bellevueActivityForm) hasOnlyZeroes() bool {
	if b.Breakfasts > 0 {
		return false
	}
	if b.Lunches > 0 {
		return false
	}
	if b.Dinners > 0 {
		return false
	}
	if b.Coffees > 0 {
		return false
	}
	if b.Saunas > 0 {
		return false
	}
	if b.Lectures > 0 {
		return false
	}
	if b.SnacksCHF > 0 {
		return false
	}
	return true
}

func (b *bellevueActivityForm) toModel() *models.BellevueActivity {
	return &models.BellevueActivity{
		ID:         b.ID,
		UserID:     b.UserID,
		Date:       b.Date,
		Breakfasts: b.Breakfasts,
		Lunches:    b.Lunches,
		Dinners:    b.Dinners,
		Coffees:    b.Coffees,
		Saunas:     b.Saunas,
		SnacksCHF:  b.SnacksCHF,
		Lectures:   b.Lectures,
		Comment:    b.Comment,
	}
}
