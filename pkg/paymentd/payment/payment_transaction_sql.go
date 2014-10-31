package payment

const selectPaymentTransaction = `
SELECT
	tx.project_id,
	tx.payment_id,
	tx.timestamp,

	tx.amount,
	tx.subunits,
	tx.currency,
	tx.status,
	tx.comment
FROM payment_transaction AS tx
`
