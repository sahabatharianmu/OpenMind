package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/code"
	"github.com/johnfercher/maroto/v2/pkg/components/col"
	"github.com/johnfercher/maroto/v2/pkg/components/row"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/config"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/props"
	"github.com/sahabatharianmu/OpenMind/pkg/response"
)

func (s *invoiceService) GenerateSuperbill(
	ctx context.Context,
	id uuid.UUID,
	organizationID uuid.UUID,
) ([]byte, error) {
	invoice, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if invoice.OrganizationID != organizationID {
		return nil, response.ErrNotFound
	}

	org, err := s.orgRepo.GetByID(organizationID)
	if err != nil {
		return nil, err
	}

	patient, err := s.patientRepo.FindByID(invoice.PatientID)
	if err != nil {
		return nil, err
	}

	var appointmentDate time.Time
	var cptCode string
	var icd10Code string

	if invoice.AppointmentID != nil {
		appt, err := s.appointmentRepo.FindByID(*invoice.AppointmentID)
		if err == nil {
			appointmentDate = appt.StartTime
			cptCode = appt.CPTCode

			note, err := s.clinicalNoteRepo.FindByAppointmentID(appt.ID)
			if err == nil {
				icd10Code = note.ICD10Code
			}
		}
	}

	m := maroto.New(config.NewBuilder().Build())

	// Header
	m.AddRows(
		row.New(20).Add(
			col.New(12).Add(
				text.New("SUPERBILL / RECEIPT", props.Text{
					Top:   5,
					Size:  20,
					Align: align.Center,
					Style: fontstyle.Bold,
				}),
			),
		),
	)

	// Practice & Patient Info
	m.AddRows(
		row.New(10).Add(
			col.New(6).Add(text.New("PRACTICE INFORMATION", props.Text{Style: fontstyle.Bold})),
			col.New(6).Add(text.New("PATIENT INFORMATION", props.Text{Style: fontstyle.Bold})),
		),
		row.New(20).Add(
			col.New(6).Add(
				text.New(org.Name),
				text.New(org.Address, props.Text{Top: 5}),
				text.New(fmt.Sprintf("Tax ID: %s", org.TaxID), props.Text{Top: 10}),
				text.New(fmt.Sprintf("NPI: %s", org.NPI), props.Text{Top: 15}),
			),
			col.New(6).Add(
				text.New(fmt.Sprintf("%s %s", patient.FirstName, patient.LastName)),
				text.New(fmt.Sprintf("DOB: %s", patient.DateOfBirth.Format("2006-01-02")), props.Text{Top: 5}),
				text.New(fmt.Sprintf("ID: %s", patient.ID.String()[:8]), props.Text{Top: 10}),
			),
		),
	)

	// Service Details
	m.AddRows(
		row.New(10).Add(
			col.New(12).Add(text.New("SERVICE DETAILS", props.Text{Style: fontstyle.Bold, Top: 10})),
		),
	)

	tableHead := row.New(10).Add(
		col.New(3).Add(text.New("Date", props.Text{Style: fontstyle.Bold})),
		col.New(3).Add(text.New("CPT Code", props.Text{Style: fontstyle.Bold})),
		col.New(3).Add(text.New("ICD-10", props.Text{Style: fontstyle.Bold})),
		col.New(3).Add(text.New("Amount", props.Text{Style: fontstyle.Bold, Align: align.Right})),
	)

	var dateStr string
	if !appointmentDate.IsZero() {
		dateStr = appointmentDate.Format("2006-01-02")
	} else {
		dateStr = invoice.CreatedAt.Format("2006-01-02")
	}

	tableRow := row.New(10).Add(
		col.New(3).Add(text.New(dateStr)),
		col.New(3).Add(text.New(cptCode)),
		col.New(3).Add(text.New(icd10Code)),
		col.New(3).
			Add(text.New(fmt.Sprintf("$%.2f", float64(invoice.AmountCents)/100.0), props.Text{Align: align.Right})),
	)

	m.AddRows(tableHead, tableRow)

	// Summary
	m.AddRows(
		row.New(20).Add(
			col.New(8).Add(text.New("TOTAL PAID", props.Text{Top: 10, Style: fontstyle.Bold, Align: align.Right})),
			col.New(4).
				Add(text.New(fmt.Sprintf("$%.2f", float64(invoice.AmountCents)/100.0), props.Text{Top: 10, Style: fontstyle.Bold, Align: align.Right})),
		),
	)

	// QR Code for verification (Sovereignty touch)
	m.AddRows(
		row.New(40).Add(
			col.New(12).Add(
				code.NewQr(fmt.Sprintf("OpenMind-Invoice-%s", invoice.ID.String()), props.Rect{
					Center:  true,
					Percent: 50,
				}),
			),
		),
	)

	document, err := m.Generate()
	if err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}

	return document.GetBytes(), nil
}
