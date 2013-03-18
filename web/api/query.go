package api

import (
	"code.google.com/p/gorest"
	"encoding/json"
	"errors"
	"github.com/prometheus/prometheus/rules"
	"github.com/prometheus/prometheus/rules/ast"
	"log"
	"net/http"
	"sort"
	"time"
)

func (serv MetricsService) Query(expr string, formatJson string) (result string) {
	exprNode, err := rules.LoadExprFromString(expr)
	if err != nil {
		return ast.ErrorToJSON(err)
	}

	timestamp := serv.time.Now()

	rb := serv.ResponseBuilder()
	var format ast.OutputFormat
	if formatJson != "" {
		format = ast.JSON
		rb.SetContentType(gorest.Application_Json)
	} else {
		format = ast.TEXT
		rb.SetContentType(gorest.Text_Plain)
	}

	return ast.EvalToString(exprNode, &timestamp, format)
}

func (serv MetricsService) QueryRange(expr string, end int64, duration int64, step int64) string {
	exprNode, err := rules.LoadExprFromString(expr)
	if err != nil {
		return ast.ErrorToJSON(err)
	}
	if exprNode.Type() != ast.VECTOR {
		return ast.ErrorToJSON(errors.New("Expression does not evaluate to vector type"))
	}
	rb := serv.ResponseBuilder()
	rb.SetContentType(gorest.Application_Json)

	if end == 0 {
		end = serv.time.Now().Unix()
	}

	if step < 1 {
		step = 1
	}

	if end-duration < 0 {
		duration = end
	}

	// Align the start to step "tick" boundary.
	end -= end % step

	matrix := ast.EvalVectorRange(
		exprNode.(ast.VectorNode),
		time.Unix(end-duration, 0),
		time.Unix(end, 0),
		time.Duration(step)*time.Second)

	sort.Sort(matrix)
	return ast.TypedValueToJSON(matrix, "matrix")
}

func (serv MetricsService) Metrics() string {
	metricNames, err := serv.persistence.GetAllMetricNames()
	rb := serv.ResponseBuilder()
	rb.SetContentType(gorest.Application_Json)
	if err != nil {
		log.Printf("Error loading metric names: %v", err)
		rb.SetResponseCode(http.StatusInternalServerError)
		return err.Error()
	}
	sort.Strings(metricNames)
	resultBytes, err := json.Marshal(metricNames)
	if err != nil {
		log.Printf("Error marshalling metric names: %v", err)
		rb.SetResponseCode(http.StatusInternalServerError)
		return err.Error()
	}
	return string(resultBytes)
}
