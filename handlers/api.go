package handlers

import (
	"encoding/json"
	model2 "investment-analysis/persistence/model"
	"net/http"
	"strconv"
	"strings"

	"investment-analysis/persistence"
)

func RegisterAPI(mux *http.ServeMux, db persistence.Store) {
	mux.HandleFunc("/api/investments", withCORS(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			investments, err := db.ListInvestments()
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
				return
			}
			writeJSON(w, http.StatusOK, investments)
		case http.MethodPost:
			var investment model2.Investment
			if err := json.NewDecoder(r.Body).Decode(&investment); err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
				return
			}
			if strings.TrimSpace(investment.Name) == "" || strings.TrimSpace(investment.Thesis) == "" || strings.TrimSpace(investment.AssetClass) == "" {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name, assetClass, and thesis are required"})
				return
			}
			created, err := db.CreateInvestment(investment)
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
				return
			}
			writeJSON(w, http.StatusCreated, created)
		case http.MethodOptions:
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))

	mux.HandleFunc("/api/investments/", withCORS(func(w http.ResponseWriter, r *http.Request) {
		path := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/investments/"), "/")
		if path == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing investment path"})
			return
		}

		pathParts := strings.Split(path, "/")
		investmentUUID := pathParts[0]

		if len(pathParts) == 2 && pathParts[1] == "expenses" {
			switch r.Method {
			case http.MethodGet:
				expenses, err := db.ListInvestmentExpenses(investmentUUID)
				if err != nil {
					writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
					return
				}
				writeJSON(w, http.StatusOK, expenses)
			case http.MethodPost:
				var expense model2.InvestmentExpense
				if err := json.NewDecoder(r.Body).Decode(&expense); err != nil {
					writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
					return
				}
				if strings.TrimSpace(expense.Label) == "" || strings.TrimSpace(expense.StartDate) == "" || strings.TrimSpace(expense.EndDate) == "" {
					writeJSON(w, http.StatusBadRequest, map[string]string{"error": "label, startDate, and endDate are required"})
					return
				}
				if expense.EventType == "" {
					expense.EventType = "cash-flow"
				}
				if expense.EventType != "cash-flow" && expense.EventType != "deferred-tax" {
					writeJSON(w, http.StatusBadRequest, map[string]string{"error": "eventType must be cash-flow or deferred-tax"})
					return
				}
				if expense.FlowType == "" {
					expense.FlowType = "one-time"
				}
				if expense.FlowType != "one-time" && expense.FlowType != "recurring" {
					writeJSON(w, http.StatusBadRequest, map[string]string{"error": "flowType must be one-time or recurring"})
					return
				}
				if expense.FlowType == "recurring" {
					if expense.RecurrenceInterval == "" {
						writeJSON(w, http.StatusBadRequest, map[string]string{"error": "recurrenceInterval is required for recurring flows"})
						return
					}
					if expense.RecurrenceInterval != "daily" && expense.RecurrenceInterval != "weekly" && expense.RecurrenceInterval != "monthly" && expense.RecurrenceInterval != "annually" {
						writeJSON(w, http.StatusBadRequest, map[string]string{"error": "recurrenceInterval must be daily, weekly, monthly, or annually"})
						return
					}
					if expense.AdjustmentFrequency != "" && expense.AdjustmentFrequency != "month" && expense.AdjustmentFrequency != "year" {
						writeJSON(w, http.StatusBadRequest, map[string]string{"error": "adjustmentFrequency must be month or year"})
						return
					}
					if expense.AdjustmentMode != "" && expense.AdjustmentMode != "amount" && expense.AdjustmentMode != "percentage" {
						writeJSON(w, http.StatusBadRequest, map[string]string{"error": "adjustmentMode must be amount or percentage"})
						return
					}
				} else {
					expense.RecurrenceInterval = ""
					expense.AdjustmentFrequency = ""
					expense.AdjustmentMode = ""
					expense.AdjustmentValue = 0
				}
				expense.DueDate = expense.EndDate
				if _, err := db.GetInvestment(investmentUUID); err == persistence.ErrNotFound {
					writeJSON(w, http.StatusNotFound, map[string]string{"error": "investment not found"})
					return
				} else if err != nil {
					writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
					return
				}
				created, err := db.CreateInvestmentExpense(investmentUUID, expense)
				if err != nil {
					writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
					return
				}
				writeJSON(w, http.StatusCreated, created)
			case http.MethodOptions:
				w.WriteHeader(http.StatusNoContent)
			default:
				w.WriteHeader(http.StatusMethodNotAllowed)
			}
			return
		}

		if len(pathParts) == 2 && pathParts[1] == "sale-assumptions" {
			switch r.Method {
			case http.MethodGet:
				assumptions, err := db.ListInvestmentSaleAssumptions(investmentUUID)
				if err != nil {
					writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
					return
				}
				writeJSON(w, http.StatusOK, assumptions)
			case http.MethodPost:
				var assumption model2.InvestmentSaleAssumption
				if err := json.NewDecoder(r.Body).Decode(&assumption); err != nil {
					writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
					return
				}
				if strings.TrimSpace(assumption.Label) == "" {
					writeJSON(w, http.StatusBadRequest, map[string]string{"error": "label is required"})
					return
				}
				if assumption.GrowthType == "" {
					assumption.GrowthType = "fixed"
				}
				if assumption.GrowthType != "fixed" && assumption.GrowthType != "growing" {
					writeJSON(w, http.StatusBadRequest, map[string]string{"error": "growthType must be fixed or growing"})
					return
				}
				if assumption.GrowthType == "growing" {
					if assumption.GrowthPeriod != "month" && assumption.GrowthPeriod != "year" {
						writeJSON(w, http.StatusBadRequest, map[string]string{"error": "growthPeriod must be month or year"})
						return
					}
					if assumption.GrowthMode != "amount" && assumption.GrowthMode != "percentage" {
						writeJSON(w, http.StatusBadRequest, map[string]string{"error": "growthMode must be amount or percentage"})
						return
					}
				} else {
					assumption.GrowthPeriod = ""
					assumption.GrowthMode = ""
					assumption.GrowthValue = 0
				}
				created, err := db.CreateInvestmentSaleAssumption(investmentUUID, assumption)
				if err != nil {
					writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
					return
				}
				writeJSON(w, http.StatusCreated, created)
			case http.MethodOptions:
				w.WriteHeader(http.StatusNoContent)
			default:
				w.WriteHeader(http.StatusMethodNotAllowed)
			}
			return
		}

		if len(pathParts) == 2 && pathParts[1] == "categories" {
			switch r.Method {
			case http.MethodGet:
				categories, err := db.ListInvestmentCategories(investmentUUID)
				if err != nil {
					writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
					return
				}
				writeJSON(w, http.StatusOK, categories)
			case http.MethodPost:
				var payload struct {
					Name string `json:"name"`
				}
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
					return
				}
				if strings.TrimSpace(payload.Name) == "" || strings.EqualFold(strings.TrimSpace(payload.Name), "uncategorized") {
					writeJSON(w, http.StatusBadRequest, map[string]string{"error": "category name is required"})
					return
				}
				if err := db.CreateInvestmentCategory(investmentUUID, payload.Name); err != nil {
					writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
					return
				}
				categories, err := db.ListInvestmentCategories(investmentUUID)
				if err != nil {
					writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
					return
				}
				writeJSON(w, http.StatusCreated, categories)
			case http.MethodOptions:
				w.WriteHeader(http.StatusNoContent)
			default:
				w.WriteHeader(http.StatusMethodNotAllowed)
			}
			return
		}

		if len(pathParts) == 3 && pathParts[1] == "categories" {
			categoryName := pathParts[2]
			if categoryName == "" {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing category name"})
				return
			}

			switch r.Method {
			case http.MethodPut:
				var payload struct {
					Name string `json:"name"`
				}
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
					return
				}
				if strings.TrimSpace(payload.Name) == "" || strings.EqualFold(strings.TrimSpace(payload.Name), "uncategorized") {
					writeJSON(w, http.StatusBadRequest, map[string]string{"error": "category name is required"})
					return
				}
				if err := db.UpdateInvestmentCategory(investmentUUID, categoryName, payload.Name); err == persistence.ErrNotFound {
					writeJSON(w, http.StatusNotFound, map[string]string{"error": "category not found"})
					return
				} else if err != nil {
					writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
					return
				}
				categories, err := db.ListInvestmentCategories(investmentUUID)
				if err != nil {
					writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
					return
				}
				writeJSON(w, http.StatusOK, categories)
			case http.MethodDelete:
				if err := db.DeleteInvestmentCategory(investmentUUID, categoryName); err == persistence.ErrNotFound {
					writeJSON(w, http.StatusNotFound, map[string]string{"error": "category not found"})
					return
				} else if err != nil {
					writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
					return
				}
				categories, err := db.ListInvestmentCategories(investmentUUID)
				if err != nil {
					writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
					return
				}
				writeJSON(w, http.StatusOK, categories)
			case http.MethodOptions:
				w.WriteHeader(http.StatusNoContent)
			default:
				w.WriteHeader(http.StatusMethodNotAllowed)
			}
			return
		}

		if len(pathParts) == 3 && pathParts[1] == "expenses" {
			expenseID, err := strconv.ParseInt(pathParts[2], 10, 64)
			if err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid expense id"})
				return
			}

			switch r.Method {
			case http.MethodPut:
				var expense model2.InvestmentExpense
				if err := json.NewDecoder(r.Body).Decode(&expense); err != nil {
					writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
					return
				}
				if strings.TrimSpace(expense.Label) == "" || strings.TrimSpace(expense.StartDate) == "" || strings.TrimSpace(expense.EndDate) == "" {
					writeJSON(w, http.StatusBadRequest, map[string]string{"error": "label, startDate, and endDate are required"})
					return
				}
				if expense.EventType == "" {
					expense.EventType = "cash-flow"
				}
				if expense.EventType != "cash-flow" && expense.EventType != "deferred-tax" {
					writeJSON(w, http.StatusBadRequest, map[string]string{"error": "eventType must be cash-flow or deferred-tax"})
					return
				}
				if expense.FlowType == "" {
					expense.FlowType = "one-time"
				}
				if expense.FlowType != "one-time" && expense.FlowType != "recurring" {
					writeJSON(w, http.StatusBadRequest, map[string]string{"error": "flowType must be one-time or recurring"})
					return
				}
				if expense.FlowType == "recurring" {
					if expense.RecurrenceInterval == "" {
						writeJSON(w, http.StatusBadRequest, map[string]string{"error": "recurrenceInterval is required for recurring flows"})
						return
					}
					if expense.RecurrenceInterval != "daily" && expense.RecurrenceInterval != "weekly" && expense.RecurrenceInterval != "monthly" && expense.RecurrenceInterval != "annually" {
						writeJSON(w, http.StatusBadRequest, map[string]string{"error": "recurrenceInterval must be daily, weekly, monthly, or annually"})
						return
					}
					if expense.AdjustmentFrequency != "" && expense.AdjustmentFrequency != "month" && expense.AdjustmentFrequency != "year" {
						writeJSON(w, http.StatusBadRequest, map[string]string{"error": "adjustmentFrequency must be month or year"})
						return
					}
					if expense.AdjustmentMode != "" && expense.AdjustmentMode != "amount" && expense.AdjustmentMode != "percentage" {
						writeJSON(w, http.StatusBadRequest, map[string]string{"error": "adjustmentMode must be amount or percentage"})
						return
					}
				} else {
					expense.RecurrenceInterval = ""
					expense.AdjustmentFrequency = ""
					expense.AdjustmentMode = ""
					expense.AdjustmentValue = 0
				}
				expense.DueDate = expense.EndDate
				updated, err := db.UpdateInvestmentExpense(investmentUUID, expenseID, expense)
				if err == persistence.ErrNotFound {
					writeJSON(w, http.StatusNotFound, map[string]string{"error": "payment flow not found"})
					return
				}
				if err != nil {
					writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
					return
				}
				writeJSON(w, http.StatusOK, updated)
			case http.MethodDelete:
				err := db.DeleteInvestmentExpense(investmentUUID, expenseID)
				if err == persistence.ErrNotFound {
					writeJSON(w, http.StatusNotFound, map[string]string{"error": "payment flow not found"})
					return
				}
				if err != nil {
					writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
					return
				}
				writeJSON(w, http.StatusOK, map[string]any{"id": expenseID, "deleted": true})
			case http.MethodOptions:
				w.WriteHeader(http.StatusNoContent)
			default:
				w.WriteHeader(http.StatusMethodNotAllowed)
			}
			return
		}

		if len(pathParts) == 3 && pathParts[1] == "sale-assumptions" {
			assumptionID, err := strconv.ParseInt(pathParts[2], 10, 64)
			if err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid sale assumption id"})
				return
			}

			switch r.Method {
			case http.MethodPut:
				var assumption model2.InvestmentSaleAssumption
				if err := json.NewDecoder(r.Body).Decode(&assumption); err != nil {
					writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
					return
				}
				if strings.TrimSpace(assumption.Label) == "" {
					writeJSON(w, http.StatusBadRequest, map[string]string{"error": "label is required"})
					return
				}
				if assumption.GrowthType == "" {
					assumption.GrowthType = "fixed"
				}
				if assumption.GrowthType != "fixed" && assumption.GrowthType != "growing" {
					writeJSON(w, http.StatusBadRequest, map[string]string{"error": "growthType must be fixed or growing"})
					return
				}
				if assumption.GrowthType == "growing" {
					if assumption.GrowthPeriod != "month" && assumption.GrowthPeriod != "year" {
						writeJSON(w, http.StatusBadRequest, map[string]string{"error": "growthPeriod must be month or year"})
						return
					}
					if assumption.GrowthMode != "amount" && assumption.GrowthMode != "percentage" {
						writeJSON(w, http.StatusBadRequest, map[string]string{"error": "growthMode must be amount or percentage"})
						return
					}
				} else {
					assumption.GrowthPeriod = ""
					assumption.GrowthMode = ""
					assumption.GrowthValue = 0
				}
				updated, err := db.UpdateInvestmentSaleAssumption(investmentUUID, assumptionID, assumption)
				if err == persistence.ErrNotFound {
					writeJSON(w, http.StatusNotFound, map[string]string{"error": "sale assumption not found"})
					return
				}
				if err != nil {
					writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
					return
				}
				writeJSON(w, http.StatusOK, updated)
			case http.MethodDelete:
				err := db.DeleteInvestmentSaleAssumption(investmentUUID, assumptionID)
				if err == persistence.ErrNotFound {
					writeJSON(w, http.StatusNotFound, map[string]string{"error": "sale assumption not found"})
					return
				}
				if err != nil {
					writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
					return
				}
				writeJSON(w, http.StatusOK, map[string]any{"id": assumptionID, "deleted": true})
			case http.MethodOptions:
				w.WriteHeader(http.StatusNoContent)
			default:
				w.WriteHeader(http.StatusMethodNotAllowed)
			}
			return
		}

		if len(pathParts) > 1 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid investment path"})
			return
		}

		switch r.Method {
		case http.MethodGet:
			investment, err := db.GetInvestment(investmentUUID)
			if err == persistence.ErrNotFound {
				writeJSON(w, http.StatusNotFound, map[string]string{"error": "investment not found"})
				return
			}
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
				return
			}
			writeJSON(w, http.StatusOK, investment)
		case http.MethodPut:
			var investment model2.Investment
			if err := json.NewDecoder(r.Body).Decode(&investment); err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
				return
			}
			if strings.TrimSpace(investment.Name) == "" || strings.TrimSpace(investment.Thesis) == "" || strings.TrimSpace(investment.AssetClass) == "" {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name, assetClass, and thesis are required"})
				return
			}
			updated, err := db.UpdateInvestment(investmentUUID, investment)
			if err == persistence.ErrNotFound {
				writeJSON(w, http.StatusNotFound, map[string]string{"error": "investment not found"})
				return
			}
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
				return
			}
			writeJSON(w, http.StatusOK, updated)
		case http.MethodOptions:
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))

	mux.HandleFunc("/api/settings", withCORS(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			settings, err := db.ListSettings()
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
				return
			}
			writeJSON(w, http.StatusOK, settings)
		case http.MethodOptions:
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))

	mux.HandleFunc("/api/settings/", withCORS(func(w http.ResponseWriter, r *http.Request) {
		provider := strings.TrimPrefix(r.URL.Path, "/api/settings/")
		provider = strings.Trim(provider, "/")
		if provider == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing provider"})
			return
		}

		switch r.Method {
		case http.MethodGet:
			settings, err := db.GetSettings(provider)
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
				return
			}
			writeJSON(w, http.StatusOK, settings)
		case http.MethodPut, http.MethodPost:
			var s model2.LLMSettings
			if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
				return
			}
			s.Provider = provider
			if s.Endpoint == "" || s.Model == "" {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "endpoint and model are required"})
				return
			}
			if err := db.UpsertSettings(s); err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
				return
			}
			writeJSON(w, http.StatusOK, s)
		case http.MethodDelete:
			if err := db.DeleteSettings(provider); err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{"provider": provider, "deleted": true, "defaults": model2.DefaultSettings(provider)})
		case http.MethodOptions:
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))
}

func withCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		next(w, r)
	}
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
