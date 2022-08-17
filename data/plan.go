package data

import (
	"context"
	"fmt"
	"log"
	"time"
)

type Plan struct {
	ID         int
	PlanName   string
	PlanAmount int
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (p *Plan) GetAll() ([]*Plan, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `select id, plan_name, plan_amount created_at, updated_at
	from plans order by plan_name`

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var plans []*Plan

	for rows.Next() {
		var plan Plan
		err := rows.Scan(
			&plan.ID,
			&plan.PlanName,
			&plan.PlanAmount,
			&plan.CreatedAt,
			&plan.UpdatedAt,
		)
		if err != nil {
			log.Println("Error scanning", err)
			return nil, err
		}

		plans = append(plans, &plan)
	}

	return plans, nil
}

func (p *Plan) GetOne(id int) (*Plan, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `select id, plan_name, plan_amount, created_at, updated_at from plans where id = $1`

	var plan Plan
	row := db.QueryRowContext(ctx, query, id)

	err := row.Scan(
		&plan.ID,
		&plan.PlanName,
		&plan.PlanAmount,
		&plan.CreatedAt,
		&plan.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &plan, nil
}

func (p *Plan) SubscribeUserToPlan(user User, plan Plan) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	stmt := `insert into user_plans (user_id, plan_id, created_at, updated_at)
			values ($1, $2, $3, $4)`

	_, err := db.ExecContext(ctx, stmt, user.ID, plan.ID, time.Now(), time.Now())
	if err != nil {
		return err
	}
	return nil
}

func (p *Plan) AmountForDisplay() string {
	amount := float64(p.PlanAmount) / 100.0
	return fmt.Sprintf("$%.2f", amount)
}
