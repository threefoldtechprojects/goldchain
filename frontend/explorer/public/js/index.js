function fillGeneralStats() {
	var request = new XMLHttpRequest();
	request.open('GET', '/explorer', true);
	request.onload = function() {
		if (request.status !== 200) {
			return;
		}
		var explorerStatus = JSON.parse(request.responseText);

		var height = document.getElementById('height');
		linkHeight(height, explorerStatus.height);

		var blockID = document.getElementById('blockID');
		linkHash(blockID, explorerStatus.blockid);

		document.getElementById('difficulty').innerHTML = readableDifficulty(explorerStatus.difficulty);
	// 	document.getElementById('maturityTimestamp').innerHTML = formatUnixTime(explorerStatus.maturitytimestamp);
	// 	document.getElementById('totalCoins').innerHTML = readableCoins(explorerStatus.totalcoins);
 	};
	request.send();
}
function fillCoinOutputStats() {
	var request = new XMLHttpRequest();
	request.open('GET', '/explorer/custodyfees/metrics/chain', true);
	request.onload = function() {
		if (request.status !== 200) {
			return;
		}
		var chainStats = JSON.parse(request.responseText);

		document.getElementById('time').innerHTML = formatUnixTime(chainStats.time);

		document.getElementById('tokensSpendable').innerHTML = readableCoins(chainStats.spendabletokens);
		document.getElementById('tokensLocked').innerHTML = readableCoins(chainStats.spendablelockedtokens);
		document.getElementById('custodyFeeDebt').innerHTML = readableCoins(chainStats.totalcustodyfeedebt);

		document.getElementById('tokensSpent').innerHTML = readableCoins(chainStats.spenttokens);
		document.getElementById('custodyFeeCollected').innerHTML = readableCoins(chainStats.paidcustodyfees);
 	};
	request.send();
}
fillGeneralStats();
fillCoinOutputStats();
