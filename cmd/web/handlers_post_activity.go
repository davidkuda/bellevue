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

// POST /activities
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

	activity := formNew.toActivity(userID)
	activityID, err := app.models.Activities.InsertWithTransaction(activity, tx)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	consumptions := formNew.toConsumptions(activityID, app.productIDMap)
	err = app.models.Consumptions.InsertManyWithTransaction(activityID, consumptions, tx)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	if err := tx.Commit(); err != nil {
		app.serverError(w, r, fmt.Errorf("failed committing transaction: %s", err))
		return
	}

	// TODO: send some notification (Toast) to the UI (successfully submitted)

	app.getActivities(w, r)
}

// PUT /activities/{id}
func (app *application) putActivitiesID(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Printf("Failed parsing form: %v", err)
		app.renderClientError(w, r, http.StatusBadRequest)
		return
	}

	activityIDString := r.PathValue("id")
	activityID, err := strconv.Atoi(activityIDString)
	if err != nil {
		app.serverError(w, r, fmt.Errorf("invalid activityID in path, could not parse: %v", err))
		return
	}

	// TODO: define one way to get this and remove the TODO comments ...
	user := app.contextGetUser(r)
	userID := user.ID

	productForm := app.parseProductForm(r)
	productForm.UserID = userID

	// TODO: if ValidationErrors, return form with errors
	if len(productForm.FieldErrors) > 0 {
		t := app.newTemplateData(r)
		app.render(w, r, http.StatusUnprocessableEntity, "activities.new.tmpl.html", &t)
		return
	}

	ctx := context.TODO()
	tx, err := app.db.BeginTx(ctx, nil)
	if err != nil {
		app.serverError(w, r, fmt.Errorf("failed starting transaction: %e", err))
		return
	}
	defer tx.Rollback()

	activity := productForm.toActivity(userID)
	activity.ID = activityID
	err = app.models.Activities.UpdateDateAndCommentTx(activity, tx)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	consumptions := productForm.toConsumptions(activityID, app.productIDMap)
	err = app.models.Consumptions.InsertManyWithTransaction(activityID, consumptions, tx)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	if err := tx.Commit(); err != nil {
		app.serverError(w, r, fmt.Errorf("failed committing transaction: %s", err))
		return
	}

	// TODO: send some notification (Toast) to the UI (successfully submitted)
	// Akshually, what I would prefer is to highlight the consumption that
	// was just created or updated and make sure it's in the viewport.

	app.getActivities(w, r)
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

func (p *productForm) toActivity(userID int) *models.Activity {
	var comm sql.NullString
	if p.Comment == "" {
		comm = sql.NullString{Valid: false}
	} else {
		comm = sql.NullString{String: p.Comment, Valid: true}
	}
	return &models.Activity{
		UserID:  userID,
		Date:    p.Date,
		Comment: comm,
	}
}

func (pf *productForm) toConsumptions(activityID int, productIDMap map[string]int) []models.Consumption {
	var consumptions models.Consumptions
	for _, p := range pf.Products {
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
			ProductID:  productIDMap[p.Code+pricecat],
			UnitPrice:  price,
			Quantity:   p.Quantity,
		}
		consumptions = append(consumptions, consumption)
	}

	return consumptions
}
