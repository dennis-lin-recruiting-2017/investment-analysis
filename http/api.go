package http

import (
	"encoding/json"
	model2 "investment-analysis/persistence/model"
	nethttp "net/http"
	"strconv"
	"strings"

	"investment-analysis/persistence"
)

func RegisterAPI(mux *nethttp.ServeMux, db persistence.Store) {
	mux.HandleFunc("/api/investments", withCORS(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		switch r.Method {
		case nethttp.MethodGet:
			investments, err := db.ListInvestments()
			if err != nil {
				writeJSON(w, nethttp.StatusInternalServerError, map[string]string{"error": err.Error()})
				return
			}
			writeJSON(w, nethttp.StatusOK, investments)
		case nethttp.MethodPost:
			var investment model2.Investment
			if err := json.NewDecoder(r.Body).Decode(&investment); err != nil {
				writeJSON(w, nethttp.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
				return
			}
			if strings.TrimSpace(investment.Name) == "" || strings.TrimSpace(investment.Thesis) == "" || strings.TrimSpace(investment.AssetClass) == "" {
				writeJSON(w, nethttp.StatusBadRequest, map[string]string{"error": "name, assetClass, and thesis are required"})
				return
			}
			created, err := db.CreateInvestment(investment)
			if err != nil {
				writeJSON(w, nethttp.StatusInternalServerError, map[string]string{"error": err.Error()})
				return
			}
			writeJSON(w, nethttp.StatusCreated, created)
		case nethttp.MethodOptions:
			w.WriteHeader(nethttp.StatusNoContent)
		default:
			w.WriteHeader(nethttp.StatusMethodNotAllowed)
		}
	}))

	mux.HandleFunc("/api/investments/", withCORS(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		path := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/investments/"), "/")
		if path == "" {
			writeJSON(w, nethttp.StatusBadRequest, map[string]string{"error": "missing investment path"})
			return
		}

		pathParts := strings.Split(path, "/")
		investmentUUID := pathParts[0]

		if len(pathParts) == 2 && pathParts[1] == "expenses" {
			switch r.Method {
			case nethttp.MethodGet:
				expenses, err := db.ListInvestmentExpenses(investmentUUID)
				if err != nil {
					writeJSON(w, nethttp.StatusInternalServerError, map[string]string{"error": err.Error()})
					return
				}
				writeJSON(w, nethttp.StatusOK, expenses)
			case nethttp.MethodPost:
				var expense model2.InvestmentExpense
				if err := json.NewDecoder(r.Body).Decode(&expense); err != nil {
					writeJSON(w, nethttp.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
					return
				}
				if strings.TrimSpace(expense.Label) == "" || strings.TrimSpace(expense.StartDate) == "" || strings.TrimSpace(expense.EndDate) == "" {
					writeJSON(w, nethttp.StatusBadRequest, map[string]string{"error": "label, startDate, and endDate are required"})
					return
				}
				if expense.EventType == "" {
					expense.EventType = "cash-flow"
				}
				if expense.EventType != "cash-flow" && expense.EventType != "deferred-tax" {
					writeJSON(w, nethttp.StatusBadRequest, map[string]string{"error": "eventType must be cash-flow or deferred-tax"})
					return
				}
				if expense.FlowType == "" {
					expense.FlowType = "one-time"
				}
				if expense.FlowType != "one-time" && expense.FlowType != "recurring" {
					writeJSON(w, nethttp.StatusBadRequest, map[string]string{"error": "flowType must be one-time or recurring"})
					return
				}
				if expense.FlowType == "recurring" {
					if expense.RecurrenceInterval == "" {
						writeJSON(w, nethttp.StatusBadRequest, map[string]string{"error": "recurrenceInterval is required for recurring flows"})
						return
					}
					if expense.RecurrenceInterval != "daily" && expense.RecurrenceInterval != "weekly" && expense.RecurrenceInterval != "monthly" && expense.RecurrenceInterval != "annually" {
						writeJSON(w, nethttp.StatusBadRequest, map[string]string{"error": "recurrenceInterval must be daily, weekly, monthly, or annually"})
						return
					}
					if expense.AdjustmentFrequency != "" && expense.AdjustmentFrequency != "month" && expense.AdjustmentFrequency != "year" {
						writeJSON(w, nethttp.StatusBadRequest, map[string]string{"error": "adjustmentFrequency must be month or year"})
						return
					}
					if expense.AdjustmentMode != "" && expense.AdjustmentMode != "amount" && expense.AdjustmentMode != "percentage" {
						writeJSON(w, nethttp.StatusBadRequest, map[string]string{"error": "adjustmentMode must be amount or percentage"})
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
					writeJSON(w, nethttp.StatusNotFound, map[string]string{"error": "investment not found"})
					return
				} else if err != nil {
					writeJSON(w, nethttp.StatusInternalServerError, map[string]string{"error": err.Error()})
					return
				}
				created, err := db.CreateInvestmentExpense(investmentUUID, expense)
				if err != nil {
					writeJSON(w, nethttp.StatusInternalServerError, map[string]string{"error": err.Error()})
					return
				}
				writeJSON(w, nethttp.StatusCreated, created)
			case nethttp.MethodOptions:
				w.WriteHeader(nethttp.StatusNoContent)
			default:
				w.WriteHeader(nethttp.StatusMethodNotAllowed)
			}
			return
		}

		if len(pathParts) == 2 && pathParts[1] == "sale-assumptions" {
			switch r.Method {
			case nethttp.MethodGet:
				assumptions, err := db.ListInvestmentSaleAssumptions(investmentUUID)
				if err != nil {
					writeJSON(w, nethttp.StatusInternalServerError, map[string]string{"error": err.Error()})
					return
				}
				writeJSON(w, nethttp.StatusOK, assumptions)
			case nethttp.MethodPost:
				var assumption model2.InvestmentSaleAssumption
				if err := json.NewDecoder(r.Body).Decode(&assumption); err != nil {
					writeJSON(w, nethttp.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
					return
				}
				if strings.TrimSpace(assumption.Label) == "" {
					writeJSON(w, nethttp.StatusBadRequest, map[string]string{"error": "label is required"})
					return
				}
				if assumption.GrowthType == "" {
					assumption.GrowthType = "fixed"
				}
				if assumption.GrowthType != "fixed" && assumption.GrowthType != "growing" {
					writeJSON(w, nethttp.StatusBadRequest, map[string]string{"error": "growthType must be fixed or growing"})
					return
				}
				if assumption.GrowthType == "growing" {
					if assumption.GrowthPeriod != "month" && assumption.GrowthPeriod != "year" {
						writeJSON(w, nethttp.StatusBadRequest, map[string]string{"error": "growthPeriod must be month or year"})
						return
					}
					if assumption.GrowthMode != "amount" && assumption.GrowthMode != "percentage" {
						writeJSON(w, nethttp.StatusBadRequest, map[string]string{"error": "growthMode must be amount or percentage"})
						return
					}
				} else {
					assumption.GrowthPeriod = ""
					assumption.GrowthMode = ""
					assumption.GrowthValue = 0
				}
				created, err := db.CreateInvestmentSaleAssumption(investmentUUID, assumption)
				if err != nil {
					writeJSON(w, nethttp.StatusInternalServerError, map[string]string{"error": err.Error()})
					return
				}
				writeJSON(w, nethttp.StatusCreated, created)
			case nethttp.MethodOptions:
				w.WriteHeader(nethttp.StatusNoContent)
			default:
				w.WriteHeader(nethttp.StatusMethodNotAllowed)
			}
			return
		}

		if len(pathParts) == 2 && pathParts[1] == "categories" {
			switch r.Method {
			case nethttp.MethodGet:
				categories, err := db.ListInvestmentCategories(investmentUUID)
				if err != nil {
					writeJSON(w, nethttp.StatusInternalServerError, map[string]string{"error": err.Error()})
					return
				}
				writeJSON(w, nethttp.StatusOK, categories)
			case nethttp.MethodPost:
				var payload struct {
					Name string `json:"name"`
				}
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					writeJSON(w, nethttp.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
					return
				}
				if strings.TrimSpace(payload.Name) == "" || strings.EqualFold(strings.TrimSpace(payload.Name), "uncategorized") {
					writeJSON(w, nethttp.StatusBadRequest, map[string]string{"error": "category name is required"})
					return
				}
				if err := db.CreateInvestmentCategory(investmentUUID, payload.Name); err != nil {
					writeJSON(w, nethttp.StatusInternalServerError, map[string]string{"error": err.Error()})
					return
				}
				categories, err := db.ListInvestmentCategories(investmentUUID)
				if err != nil {
					writeJSON(w, nethttp.StatusInternalServerError, map[string]string{"error": err.Error()})
					return
				}
				writeJSON(w, nethttp.StatusCreated, categories)
			case nethttp.MethodOptions:
				w.WriteHeader(nethttp.StatusNoContent)
			default:
				w.WriteHeader(nethttp.StatusMethodNotAllowed)
			}
			return
		}

		if len(pathParts) == 3 && pathParts[1] == "categories" {
			categoryName := pathParts[2]
			if categoryName == "" {
				writeJSON(w, nethttp.StatusBadRequest, map[string]string{"error": "missing category name"})
				return
			}

			switch r.Method {
			case nethttp.MethodPut:
				var payload struct {
					Name string `json:"name"`
				}
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					writeJSON(w, nethttp.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
					return
				}
				if strings.TrimSpace(payload.Name) == "" || strings.EqualFold(strings.TrimSpace(payload.Name), "uncategorized") {
					writeJSON(w, nethttp.StatusBadRequest, map[string]string{"error": "category name is required"})
					return
				}
				if err := db.UpdateInvestmentCategory(investmentUUID, categoryName, payload.Name); err == persistence.ErrNotFound {
					writeJSON(w, nethttp.StatusNotFound, map[string]string{"error": "category not found"})
					return
				} else if err != nil {
					writeJSON(w, nethttp.StatusInternalServerError, map[string]string{"error": err.Error()})
					return
				}
				categories, err := db.ListInvestmentCategories(investmentUUID)
				if err != nil {
					writeJSON(w, nethttp.StatusInternalServerError, map[string]string{"error": err.Error()})
					return
				}
				writeJSON(w, nethttp.StatusOK, categories)
			case nethttp.MethodDelete:
				if err := db.DeleteInvestmentCategory(investmentUUID, categoryName); err == persistence.ErrNotFound {
					writeJSON(w, nethttp.StatusNotFound, map[string]string{"error": "category not found"})
					return
				} else if err != nil {
					writeJSON(w, nethttp.StatusInternalServerError, map[string]string{"error": err.Error()})
					return
				}
				categories, err := db.ListInvestmentCategories(investmentUUID)
				if err != nil {
					writeJSON(w, nethttp.StatusInternalServerError, map[string]string{"error": err.Error()})
					return
				}
				writeJSON(w, nethttp.StatusOK, categories)
			case nethttp.MethodOptions:
				w.WriteHeader(nethttp.StatusNoContent)
			default:
				w.WriteHeader(nethttp.StatusMethodNotAllowed)
			}
			return
		}

		if len(pathParts) == 3 && pathParts[1] == "expenses" {
			expenseID, err := strconv.ParseInt(pathParts[2], 10, 64)
			if err != nil {
				writeJSON(w, nethttp.StatusBadRequest, map[string]string{"error": "invalid expense id"})
				return
			}

			switch r.Method {
			case nethttp.MethodPut:
				var expense model2.InvestmentExpense
				if err := json.NewDecoder(r.Body).Decode(&expense); err != nil {
					writeJSON(w, nethttp.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
					return
				}
				if strings.TrimSpace(expense.Label) == "" || strings.TrimSpace(expense.StartDate) == "" || strings.TrimSpace(expense.EndDate) == "" {
					writeJSON(w, nethttp.StatusBadRequest, map[string]string{"error": "label, startDate, and endDate are required"})
					return
				}
				if expense.EventType == "" {
					expense.EventType = "cash-flow"
				}
				if expense.EventType != "cash-flow" && expense.EventType != "deferred-tax" {
					writeJSON(w, nethttp.StatusBadRequest, map[string]string{"error": "eventType must be cash-flow or deferred-tax"})
					return
				}
				if expense.FlowType == "" {
					expense.FlowType = "one-time"
				}
				if expense.FlowType != "one-time" && expense.FlowType != "recurring" {
					writeJSON(w, nethttp.StatusBadRequest, map[string]string{"error": "flowType must be one-time or recurring"})
					return
				}
				if expense.FlowType == "recurring" {
					if expense.RecurrenceInterval == "" {
						writeJSON(w, nethttp.StatusBadRequest, map[string]string{"error": "recurrenceInterval is required for recurring flows"})
						return
					}
					if expense.RecurrenceInterval != "daily" && expense.RecurrenceInterval != "weekly" && expense.RecurrenceInterval != "monthly" && expense.RecurrenceInterval != "annually" {
						writeJSON(w, nethttp.StatusBadRequest, map[string]string{"error": "recurrenceInterval must be daily, weekly, monthly, or annually"})
						return
					}
					if expense.AdjustmentFrequency != "" && expense.AdjustmentFrequency != "month" && expense.AdjustmentFrequency != "year" {
						writeJSON(w, nethttp.StatusBadRequest, map[string]string{"error": "adjustmentFrequency must be month or year"})
						return
					}
					if expense.AdjustmentMode != "" && expense.AdjustmentMode != "amount" && expense.AdjustmentMode != "percentage" {
						writeJSON(w, nethttp.StatusBadRequest, map[string]string{"error": "adjustmentMode must be amount or percentage"})
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
					writeJSON(w, nethttp.StatusNotFound, map[string]string{"error": "payment flow not found"})
					return
				}
				if err != nil {
					writeJSON(w, nethttp.StatusInternalServerError, map[string]string{"error": err.Error()})
					return
				}
				writeJSON(w, nethttp.StatusOK, updated)
			case nethttp.MethodDelete:
				err := db.DeleteInvestmentExpense(investmentUUID, expenseID)
				if err == persistence.ErrNotFound {
					writeJSON(w, nethttp.StatusNotFound, map[string]string{"error": "payment flow not found"})
					return
				}
				if err != nil {
					writeJSON(w, nethttp.StatusInternalServerError, map[string]string{"error": err.Error()})
					return
				}
				writeJSON(w, nethttp.StatusOK, map[string]any{"id": expenseID, "deleted": true})
			case nethttp.MethodOptions:
				w.WriteHeader(nethttp.StatusNoContent)
			default:
				w.WriteHeader(nethttp.StatusMethodNotAllowed)
			}
			return
		}

		if len(pathParts) == 3 && pathParts[1] == "sale-assumptions" {
			assumptionID, err := strconv.ParseInt(pathParts[2], 10, 64)
			if err != nil {
				writeJSON(w, nethttp.StatusBadRequest, map[string]string{"error": "invalid sale assumption id"})
				return
			}

			switch r.Method {
			case nethttp.MethodPut:
				var assumption model2.InvestmentSaleAssumption
				if err := json.NewDecoder(r.Body).Decode(&assumption); err != nil {
					writeJSON(w, nethttp.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
					return
				}
				if strings.TrimSpace(assumption.Label) == "" {
					writeJSON(w, nethttp.StatusBadRequest, map[string]string{"error": "label is required"})
					return
				}
				if assumption.GrowthType == "" {
					assumption.GrowthType = "fixed"
				}
				if assumption.GrowthType != "fixed" && assumption.GrowthType != "growing" {
					writeJSON(w, nethttp.StatusBadRequest, map[string]string{"error": "growthType must be fixed or growing"})
					return
				}
				if assumption.GrowthType == "growing" {
					if assumption.GrowthPeriod != "month" && assumption.GrowthPeriod != "year" {
						writeJSON(w, nethttp.StatusBadRequest, map[string]string{"error": "growthPeriod must be month or year"})
						return
					}
					if assumption.GrowthMode != "amount" && assumption.GrowthMode != "percentage" {
						writeJSON(w, nethttp.StatusBadRequest, map[string]string{"error": "growthMode must be amount or percentage"})
						return
					}
				} else {
					assumption.GrowthPeriod = ""
					assumption.GrowthMode = ""
					assumption.GrowthValue = 0
				}
				updated, err := db.UpdateInvestmentSaleAssumption(investmentUUID, assumptionID, assumption)
				if err == persistence.ErrNotFound {
					writeJSON(w, nethttp.StatusNotFound, map[string]string{"error": "sale assumption not found"})
					return
				}
				if err != nil {
					writeJSON(w, nethttp.StatusInternalServerError, map[string]string{"error": err.Error()})
					return
				}
				writeJSON(w, nethttp.StatusOK, updated)
			case nethttp.MethodDelete:
				err := db.DeleteInvestmentSaleAssumption(investmentUUID, assumptionID)
				if err == persistence.ErrNotFound {
					writeJSON(w, nethttp.StatusNotFound, map[string]string{"error": "sale assumption not found"})
					return
				}
				if err != nil {
					writeJSON(w, nethttp.StatusInternalServerError, map[string]string{"error": err.Error()})
					return
				}
				writeJSON(w, nethttp.StatusOK, map[string]any{"id": assumptionID, "deleted": true})
			case nethttp.MethodOptions:
				w.WriteHeader(nethttp.StatusNoContent)
			default:
				w.WriteHeader(nethttp.StatusMethodNotAllowed)
			}
			return
		}

		if len(pathParts) > 1 {
			writeJSON(w, nethttp.StatusBadRequest, map[string]string{"error": "invalid investment path"})
			return
		}

		switch r.Method {
		case nethttp.MethodGet:
			investment, err := db.GetInvestment(investmentUUID)
			if err == persistence.ErrNotFound {
				writeJSON(w, nethttp.StatusNotFound, map[string]string{"error": "investment not found"})
				return
			}
			if err != nil {
				writeJSON(w, nethttp.StatusInternalServerError, map[string]string{"error": err.Error()})
				return
			}
			writeJSON(w, nethttp.StatusOK, investment)
		case nethttp.MethodPut:
			var investment model2.Investment
			if err := json.NewDecoder(r.Body).Decode(&investment); err != nil {
				writeJSON(w, nethttp.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
				return
			}
			if strings.TrimSpace(investment.Name) == "" || strings.TrimSpace(investment.Thesis) == "" || strings.TrimSpace(investment.AssetClass) == "" {
				writeJSON(w, nethttp.StatusBadRequest, map[string]string{"error": "name, assetClass, and thesis are required"})
				return
			}
			updated, err := db.UpdateInvestment(investmentUUID, investment)
			if err == persistence.ErrNotFound {
				writeJSON(w, nethttp.StatusNotFound, map[string]string{"error": "investment not found"})
				return
			}
			if err != nil {
				writeJSON(w, nethttp.StatusInternalServerError, map[string]string{"error": err.Error()})
				return
			}
			writeJSON(w, nethttp.StatusOK, updated)
		case nethttp.MethodDelete:
			err := db.DeleteInvestment(investmentUUID)
			if err == persistence.ErrNotFound {
				writeJSON(w, nethttp.StatusNotFound, map[string]string{"error": "investment not found"})
				return
			}
			if err != nil {
				writeJSON(w, nethttp.StatusInternalServerError, map[string]string{"error": err.Error()})
				return
			}
			writeJSON(w, nethttp.StatusOK, map[string]any{"uuid": investmentUUID, "deleted": true})
		case nethttp.MethodOptions:
			w.WriteHeader(nethttp.StatusNoContent)
		default:
			w.WriteHeader(nethttp.StatusMethodNotAllowed)
		}
	}))

	mux.HandleFunc("/api/settings", withCORS(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		switch r.Method {
		case nethttp.MethodGet:
			settings, err := db.ListSettings()
			if err != nil {
				writeJSON(w, nethttp.StatusInternalServerError, map[string]string{"error": err.Error()})
				return
			}
			writeJSON(w, nethttp.StatusOK, settings)
		case nethttp.MethodOptions:
			w.WriteHeader(nethttp.StatusNoContent)
		default:
			w.WriteHeader(nethttp.StatusMethodNotAllowed)
		}
	}))

	mux.HandleFunc("/api/settings/", withCORS(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		provider := strings.TrimPrefix(r.URL.Path, "/api/settings/")
		provider = strings.Trim(provider, "/")
		if provider == "" {
			writeJSON(w, nethttp.StatusBadRequest, map[string]string{"error": "missing provider"})
			return
		}

		switch r.Method {
		case nethttp.MethodGet:
			settings, err := db.GetSettings(provider)
			if err != nil {
				writeJSON(w, nethttp.StatusInternalServerError, map[string]string{"error": err.Error()})
				return
			}
			writeJSON(w, nethttp.StatusOK, settings)
		case nethttp.MethodPut, nethttp.MethodPost:
			var s model2.LLMSettings
			if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
				writeJSON(w, nethttp.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
				return
			}
			s.Provider = provider
			if s.Endpoint == "" || s.Model == "" {
				writeJSON(w, nethttp.StatusBadRequest, map[string]string{"error": "endpoint and model are required"})
				return
			}
			if err := db.UpsertSettings(s); err != nil {
				writeJSON(w, nethttp.StatusInternalServerError, map[string]string{"error": err.Error()})
				return
			}
			writeJSON(w, nethttp.StatusOK, s)
		case nethttp.MethodDelete:
			if err := db.DeleteSettings(provider); err != nil {
				writeJSON(w, nethttp.StatusInternalServerError, map[string]string{"error": err.Error()})
				return
			}
			writeJSON(w, nethttp.StatusOK, map[string]any{"provider": provider, "deleted": true, "defaults": model2.DefaultSettings(provider)})
		case nethttp.MethodOptions:
			w.WriteHeader(nethttp.StatusNoContent)
		default:
			w.WriteHeader(nethttp.StatusMethodNotAllowed)
		}
	}))

	mux.HandleFunc("/api/retrieval-settings", withCORS(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		switch r.Method {
		case nethttp.MethodGet:
			settings, err := db.GetRetrievalSettings()
			if err != nil {
				writeJSON(w, nethttp.StatusInternalServerError, map[string]string{"error": err.Error()})
				return
			}
			writeJSON(w, nethttp.StatusOK, settings)
		case nethttp.MethodPut, nethttp.MethodPost:
			var s model2.RetrievalSettings
			if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
				writeJSON(w, nethttp.StatusBadRequest, map[string]string{"error": err.Error()})
				return
			}
			if s.PlaywrightTimeoutSeconds <= 0 {
				writeJSON(w, nethttp.StatusBadRequest, map[string]string{"error": "playwrightTimeoutSeconds must be greater than 0"})
				return
			}
			if err := db.UpsertRetrievalSettings(s); err != nil {
				writeJSON(w, nethttp.StatusInternalServerError, map[string]string{"error": err.Error()})
				return
			}
			writeJSON(w, nethttp.StatusOK, s)
		case nethttp.MethodOptions:
			w.WriteHeader(nethttp.StatusNoContent)
		default:
			w.WriteHeader(nethttp.StatusMethodNotAllowed)
		}
	}))
}

func withCORS(next nethttp.HandlerFunc) nethttp.HandlerFunc {
	return func(w nethttp.ResponseWriter, r *nethttp.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		next(w, r)
	}
}

func writeJSON(w nethttp.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
