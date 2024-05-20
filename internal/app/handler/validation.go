package handler

const lastDigit = 9

func ValidateOrderNumber(number string) bool {
	var sum int
	digits := make([]int, len(number))
	for i, char := range number {
		if char < '0' || char > '9' {
			return false
		}
		digits[i] = int(char - '0')
	}

	double := false
	for i := len(digits) - 1; i >= 0; i-- {
		digit := digits[i]
		if double {
			digit *= 2
			if digit > lastDigit {
				digit -= lastDigit
			}
		}
		sum += digit
		double = !double
	}

	return sum%10 == 0
}
