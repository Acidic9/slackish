function validateDisplayName() {
	var input = $('.display-name-input');

	clearInputValidation(input);

	if (input.val().length == 0)
		return;

	$.getJSON("/api/displayNames/exists/" + input.val(), function (data) {
		if (data.exists) {
			addInputErrorMsg(input, "Display name is not available");
		} else {
			addInputSuccessMsg(input, "Display name is available");
		}
	});
}

$("#display-name-dont-ask-again").click(function () {
	$(".display-name-change").animate({
		"height": 0,
		"padding": 0
	}, function () {
		$(".display-name-change").remove();
	});
	$.post("/do/displayName/dontAskAgain",
		function (resp, status) {
			switch (status) {
				case "success":
					if (!resp.success) {
						var error = resp.error ? resp.error : "Something went wrong";
						alertify.error(error);
						return;
					}
					alertify.success("You won't be see this on the homepage again");
					break;

				case "timeout":
					alertify.error("The server didn't respond");
					break;

				default:
					alertify.error("Something went wrong");
					break;
			}
		},
		"json"
	).fail(function () {
		alertify.error("Something went wrong");
	});
});

// Searching
var lastSearch = "";
var slackSearchTypingTimer;
var slackSearchDoneTypingInterval = 260;
$("#search-slacks").keydown(function () {
	clearTimeout(slackSearchTypingTimer);
	if ($('#search-slacks').val()) {
		slackSearchTypingTimer = setTimeout(function () { searchSlacks($("#search-slacks").val()); }, slackSearchDoneTypingInterval);
	}
});
function searchSlacks(searchText) {
	if (searchText == lastSearch) return;
	if (!/[a-zA-Z0-9_]{2,}/.test(searchText)) {
		$("#slack-search-results").hide();
		$("#slack-search-results tbody").html("");
		return;
	}
	lastSearch = searchText;

	$.getJSON("/api/slacks/search/" + searchText,
		function (resp, status, jqXHR) {
			if (lastSearch != searchText) return;
			switch (status) {
				case "success":
					if (!resp.success) {
						// var error = resp.error ? resp.error : "Something went wrong";
						// alertify.error(error);
						console.log(error);
						return;
					}
					$("#slack-search-results tbody").html("");
					if (resp.slacks.length <= 0) {
						$("#slack-search-results").hide();
						return;
					}
					$.each(resp.slacks, function (i, slack) {
						$("#slack-search-results tbody").append(`<tr>
							<td><a href="/`+ slack.name + `">/` + slack.name + `</a></td>
							<td>`+ slack.description + `</td>
						</tr>`);
					});
					$("#slack-search-results").show();
					break;
			}
		}
	);
}