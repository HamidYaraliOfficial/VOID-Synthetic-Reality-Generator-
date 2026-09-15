// Package finance implements VOID's Financial Transaction Simulator: it
// generates high-volume synthetic payment/transfer/refund/subscription/
// chargeback/fraud events for testing payment backends and fraud-detection
// systems, with configurable fraud rate and pattern injection.
package finance

import (
	"fmt"

	"void/internal/rng"
)

type TxnType string

const (
	TxnPayment      TxnType = "payment"
	TxnTransfer     TxnType = "transfer"
	TxnRefund       TxnType = "refund"
	TxnSubscription TxnType = "subscription"
	TxnChargeback   TxnType = "chargeback"
)

type Transaction struct {
	ID       string
	Type     TxnType
	FromID   string
	ToID     string
	Amount   float64
	Currency string
	Tick     int64
	Fraud    bool
	FraudReason string
}

// Simulator generates synthetic financial transactions with a configurable
// baseline fraud rate plus "burst" fraud injection (e.g. via a Scenario).
type Simulator struct {
	Currency        string
	BaseFraudRate   float64
	FraudMultiplier float64
	counter         int64
}

func NewSimulator(currency string, baseFraudRate float64) *Simulator {
	return &Simulator{Currency: currency, BaseFraudRate: baseFraudRate, FraudMultiplier: 1}
}

func (s *Simulator) nextID() string {
	s.counter++
	return fmt.Sprintf("txn-%09d", s.counter)
}

// Generate creates one transaction for the given accounts/tick, deciding
// type + amount + fraud flag from the simulator's configured distributions.
func (s *Simulator) Generate(fromID, toID string, tick int64, r *rng.Source) Transaction {
	txnType := s.pickType(r)
	amount := s.pickAmount(txnType, r)
	fraudRate := s.BaseFraudRate * s.FraudMultiplier
	isFraud := r.Bool(fraudRate)
	reason := ""
	if isFraud {
		reason = pickFraudReason(r)
		amount *= r.Uniform(2, 8) // fraudulent transactions often anomalously large
	}
	return Transaction{
		ID: s.nextID(), Type: txnType, FromID: fromID, ToID: toID,
		Amount: amount, Currency: s.Currency, Tick: tick,
		Fraud: isFraud, FraudReason: reason,
	}
}

func (s *Simulator) pickType(r *rng.Source) TxnType {
	types := []TxnType{TxnPayment, TxnTransfer, TxnRefund, TxnSubscription, TxnChargeback}
	weights := []float64{0.55, 0.2, 0.1, 0.12, 0.03}
	return types[r.WeightedChoice(weights)]
}

func (s *Simulator) pickAmount(t TxnType, r *rng.Source) float64 {
	switch t {
	case TxnSubscription:
		return r.Uniform(5, 50)
	case TxnRefund, TxnChargeback:
		return r.Uniform(10, 300)
	default:
		return absf(r.Normal(80, 60))
	}
}

func pickFraudReason(r *rng.Source) string {
	reasons := []string{"stolen_card", "account_takeover", "synthetic_identity", "velocity_abuse", "chargeback_fraud"}
	return reasons[r.Choice(len(reasons))]
}

func absf(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}
