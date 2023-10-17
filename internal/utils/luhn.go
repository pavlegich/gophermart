package utils

// LuhnValid проверяет корректность номера по алгоритму Luhn
func LuhnValid(number int) bool {
	return (number%10+checksum(number/10))%10 == 0
}

// checkSum проверяет часть номера на корректность по алгоритму Luhn
func checksum(number int) int {
	var luhn int

	for i := 0; number > 0; i++ {
		cur := number % 10

		if i%2 == 0 {
			cur = cur * 2
			if cur > 9 {
				cur = cur%10 + cur/10
			}
		}

		luhn += cur
		number = number / 10
	}
	return luhn % 10
}
