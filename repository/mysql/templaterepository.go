package mysql

import (
	"fmt"
	"github.com/doug-martin/goqu/v9"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"goacs/acs/types"
	"goacs/models/cpe"
	"goacs/models/templates"
	"goacs/repository"
	"log"
	"time"
)

type TemplateRepository struct {
	db *sqlx.DB
}

func NewTemplateRepository(connection *sqlx.DB) *TemplateRepository {
	return &TemplateRepository{
		db: connection,
	}
}

func (r *TemplateRepository) Find(id int64) (*templates.Template, error) {
	templateInstance := new(templates.Template)
	err := r.db.Unsafe().Get(templateInstance, "SELECT * FROM templates WHERE id=? LIMIT 1", id)

	if err != nil {
		fmt.Println("Error while fetching query results")
		fmt.Println(err.Error())
		return nil, repository.ErrNotFound

	}

	return templateInstance, nil
}

func (r *TemplateRepository) FindByName(name string) (*templates.Template, error) {
	templateInstance := new(templates.Template)

	err := r.db.Unsafe().Get(templateInstance, "SELECT id, name FROM templates WHERE name=? LIMIT 1", name)

	if err != nil {
		fmt.Println("Error while fetching query results")
		fmt.Println(err.Error())
		return nil, repository.ErrNotFound

	}

	return templateInstance, nil
}

func (r *TemplateRepository) List(request repository.PaginatorRequest) ([]templates.Template, int) {
	var total int
	var templates = make([]templates.Template, 0)
	_ = r.db.Get(&total, "SELECT count(*) FROM templates")
	err := r.db.Unsafe().Select(&templates, "SELECT * FROM templates LIMIT ?,?", request.CalcOffset(), request.PerPage)

	if err != nil {
		fmt.Println("Error while fetching query results")
		fmt.Println(err.Error())
		return nil, 0
	}

	return templates, total
}

func (r *TemplateRepository) CreateTemplate(template *templates.Template) {
	dialect := goqu.Dialect("mysql")

	query, args, _ := dialect.Insert("templates").Prepared(true).
		Cols("name").
		Vals(goqu.Vals{
			template.Name,
		}).ToSQL()

	_, err := r.db.Exec(query, args...)

	if err != nil {
		log.Println("error CreateTemplate ", err.Error())
	}
}

func (r *TemplateRepository) GetParametersForTemplate(template_id int64) ([]templates.TemplateParameter, error) {
	var parameters = []templates.TemplateParameter{}

	err := r.db.Unsafe().Select(&parameters,
		"SELECT * FROM templates_parameters WHERE template_id=? LIMIT 1",
		template_id,
	)

	if err != nil {
		fmt.Println("Error while fetching query results")
		fmt.Println(err.Error())
		return nil, repository.ErrNotFound
	}

	return parameters, nil
}

func (r *TemplateRepository) HydrateTemplatesParameters(templatesData []templates.Template) []templates.Template {

	var parameters []templates.TemplateParameter
	var ids []int64

	for _, template := range templatesData {
		ids = append(ids, template.Id)
	}

	dialect := goqu.Dialect("mysql")
	selectSql, _, _ := dialect.From("templates_parameters").
		Where(goqu.C("template_id").In(ids)).ToSQL()

	err := r.db.Select(&parameters, selectSql)

	if err != nil {
		return templatesData
	}

	for templateIdx, template := range templatesData {
		for _, parameter := range parameters {
			if parameter.TemplateId == template.Id {
				templatesData[templateIdx].Parameters = append(templatesData[templateIdx].Parameters, parameter)
			}
		}
	}

	return templatesData
}

func (r *TemplateRepository) GetPrioritizedParametersForCPE(cpe *cpe.CPE) []types.PrioritizedParameters {
	var prioParams []types.PrioritizedParameters

	dialect := goqu.Dialect("mysql")

	orderedTemplatesIdsQuery, args, _ := dialect.From(goqu.T("templates_parameters").As("tp")).
		Prepared(true).
		Join(
			goqu.T("cpe_to_templates").As("c2t"),
			goqu.On(goqu.Ex{"c2t.template_id": goqu.I("tp.template_id")}),
		).
		Where(goqu.Ex{"c2t.cpe_uuid": cpe.UUID}).
		Order(goqu.I("c2t.priority").Asc()).ToSQL()

	//log.Println("GetPrioritizedParametersForCPE", orderedTemplatesIdsQuery)
	err := r.db.Select(&prioParams, orderedTemplatesIdsQuery, args...)

	if err != nil {
		log.Println("Error in GetPrioritizedParametersForCPE", err.Error())
	}

	return prioParams
}

func (r *TemplateRepository) ListTemplateParameters(template *templates.Template, request repository.PaginatorRequest) ([]templates.TemplateParameter, int) {
	dialect := goqu.Dialect("mysql")

	baseBulder := dialect.From("templates_parameters").
		Where(goqu.C("template_id").Eq(template.Id))

	if len(request.Filter) > 0 {
		for key, value := range request.Filter {
			baseBulder = baseBulder.Where(goqu.Ex{
				key: goqu.Op{"ilike": "%" + value + "%"},
			})
		}
	}

	totalSql, _, _ := baseBulder.
		Select(goqu.COUNT("*")).
		ToSQL()

	var total int
	_ = r.db.Get(&total, totalSql)
	var parameters []templates.TemplateParameter
	parametersBuilder := baseBulder.
		Offset(uint(request.CalcOffset())).
		Limit(uint(request.PerPage))

	log.Println(request.Filter)

	parametersSql, _, _ := parametersBuilder.ToSQL()

	log.Println(parametersSql)

	err := r.db.Unsafe().Select(&parameters, parametersSql)

	if err != nil {
		fmt.Println("Error while fetching query results")
		fmt.Println(err.Error())
		return nil, 0
	}

	return parameters, total
}

func (r *TemplateRepository) GetTemplatesForCPE(cpe *cpe.CPE) []templates.CPETemplate {
	dialect := goqu.Dialect("mysql")

	query, _, _ := dialect.From("cpe_to_templates").Join(
		goqu.T("templates").As("t"),
		goqu.On(goqu.Ex{"cpe_to_templates.template_id": goqu.I("t.id")}),
	).Where(goqu.Ex{"cpe_to_templates.cpe_uuid": cpe.UUID}).
		Order(goqu.I("priority").Desc()).ToSQL()

	var cpeTemplates []templates.CPETemplate
	err := r.db.Select(&cpeTemplates, query)
	if err != nil {
		log.Println(err)
		return []templates.CPETemplate{}
	}

	return cpeTemplates
}

func (r *TemplateRepository) AssignTemplateToDevice(cpe *cpe.CPE, template_id int64, priority int64) error {
	dialect := goqu.Dialect("mysql")

	query, args, _ := dialect.Insert("cpe_to_templates").Prepared(true).
		Cols("cpe_uuid", "template_id", "priority").
		Vals(goqu.Vals{cpe.UUID, template_id, priority}).ToSQL()

	_, err := r.db.Exec(query, args...)

	if err != nil {
		log.Println("AssignTemplateToDevice error", err)
		return err
	}

	return nil
}

func (r *TemplateRepository) UnassignTemplateFromDevice(cpe *cpe.CPE, template_id int64) error {
	dialect := goqu.Dialect("mysql")
	query, args, _ := dialect.Delete("cpe_to_templates").Prepared(true).
		Where(goqu.Ex{"cpe_uuid": cpe.UUID, "template_id": template_id}).
		ToSQL()

	_, err := r.db.Exec(query, args...)

	if err != nil {
		log.Println("UnassignTemplateFromDevice error", err)
		return err
	}

	return nil
}

func (r *TemplateRepository) FindParameterByName(template_id int64, parameter_name string) (types.ParameterValueStruct, error) {
	var templateParameter types.ParameterValueStruct

	dialect := goqu.Dialect("mysql")

	query, args, _ := dialect.From("template_parameters").
		Where(goqu.Ex{
			"template_id": template_id,
			"name":        parameter_name,
		}).ToSQL()

	err := r.db.Get(&templateParameter, query, args...)

	if err != nil {
		return templateParameter, err
	}

	return templateParameter, nil
}

func (r *TemplateRepository) CreateParameter(template_id int64, parameter types.ParameterValueStruct) error {
	uuidInstance, _ := uuid.NewRandom()
	uuidString := uuidInstance.String()

	dialect := goqu.Dialect("mysql")

	query, args, _ := dialect.Insert("templates_parameters").Prepared(true).
		Cols("uuid", "template_id", "name", "value", "type", "flags", "created_at").
		Vals(goqu.Vals{
			uuidString,
			template_id,
			parameter.Name,
			parameter.ValueStruct.Value,
			"",
			parameter.Flag.AsString(),
			time.Now(),
		}).
		ToSQL()

	_, err := r.db.Exec(query, args...)

	if err != nil {
		log.Println("Template CreateParameter error", err)
		return err
	}

	return nil
}

func (r *TemplateRepository) UpdateParameter(parameter_uuid string, parameter types.ParameterValueStruct) error {
	dialect := goqu.Dialect("mysql")

	query, args, _ := dialect.Update("templates_parameters").Prepared(true).
		Set(goqu.Record{
			"name":       parameter.Name,
			"value":      parameter.ValueStruct.Value,
			"flags":      parameter.Flag.AsString(),
			"updated_at": time.Now(),
		}).
		Where(goqu.Ex{
			"uuid": parameter_uuid,
		}).
		ToSQL()

	_, err := r.db.Exec(query, args...)

	if err != nil {
		log.Println("Template CreateParameter error", err)
		return err
	}

	return nil
}

func (r *TemplateRepository) DeleteParameter(parameter_uuid string, template_id int64) error {
	dialect := goqu.Dialect("mysql")

	query, args, _ := dialect.Delete("templates_parameters").Prepared(true).
		Where(goqu.Ex{
			"uuid":        parameter_uuid,
			"template_id": template_id,
		}).
		ToSQL()

	_, err := r.db.Exec(query, args...)

	if err != nil {
		log.Println("Template DeleteParameter error", err)
		return err
	}

	return nil
}
