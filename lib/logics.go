package lib

//CalculateLowerBound : given current_bid, current lower_bound, and bought in price, calculate new lower_bound
func CalculateLowerBound(boughtIn float64, currentBid float64, currentLowerBound float64, profitRange float64, lowerBound float64) float64 {
	if currentBid < boughtIn*profitRange && currentBid > boughtIn*lowerBound {

		return boughtIn * lowerBound

	} else if currentBid >= boughtIn*profitRange {

		if currentBid*0.99 > currentLowerBound && currentBid*0.99 >= boughtIn*profitRange {
			return currentBid * 0.99
		}
		return boughtIn * profitRange
	}

	return currentBid * lowerBound
}
