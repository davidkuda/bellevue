package main

import (
	"context"
	"database/sql"
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

	user := app.contextGetUser(r)
	userID := user.ID // TODO: Deal with case where user is nil

	formNew := app.parseProductForm(r)
	formNew.UserID = userID

	// TODO: if ValidationErrors, return form with errors
	if len(formNew.FieldErrors) > 0 {
		t := app.newTemplateData(r)
		app.render(w, r, http.StatusUnprocessableEntity, "activities.new.tmpl.html", &t)
		return
	}

	// we will do the following in a transaction:
	// - insert activity
	// - delete all previous consumptions, if any (for edits)
	//   (form upload is state of truth, remove everything else)
	// - insert all consumptions based on the form
	ctx := context.TODO()
	tx, err := app.db.BeginTx(ctx, nil)
	if err != nil {
		app.serverError(w, r, fmt.Errorf("failed starting transaction: %e", err))
		return
	}
	defer tx.Rollback()

	var comm sql.NullString
	if formNew.Comment == "" {
		comm = sql.NullString{Valid: false}
	} else {
		comm = sql.NullString{String: formNew.Comment, Valid: true}
	}
	activity := &models.Activity{
		UserID: userID,
		Date: formNew.Date,
		Comment: comm,
	}

	activityID, err := app.models.Activities.InsertWithTransaction(activity, tx)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

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
			ActivityID: activityID,
			ProductID:  app.productIDMap[p.Code+pricecat],
			UnitPrice:  price,
			Quantity:   p.Quantity,
		}
		consumptions = append(consumptions, consumption)
	}

	err = app.models.Consumptions.InsertManyWithTransaction(activityID, consumptions, tx)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	if err := tx.Commit(); err != nil {
		app.serverError(w,r, fmt.Errorf("failed committing transaction: %s", err))
		return
	}

	// TODO: send some notification (Toast) to the UI (successfully submitted)
	http.Redirect(w, r, "/activities", http.StatusSeeOther)
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
			pcid := app.priceCategoryIDMap[pricecat]
			if pcid == 0 {
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
			pp.Price = priceInt
			pp.Quantity = 1
		}
		form.Products = append(form.Products, pp)
	}

	form.Comment = r.PostForm.Get("comment")

	return form
}
