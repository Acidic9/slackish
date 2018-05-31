$(document).ready(function() {
	// Validate slack name on finish typing
	var typingTimer;
	var doneTypingInterval = 500; // In ms, how long to wait before validating slack name

	//// On keyup, start the countdown
	$('#slack-name').keyup(function() {
		clearInputValidation($(this));
		clearTimeout(typingTimer);
		if ($('#slack-name').val()) {
			typingTimer = setTimeout(validateSlackName, doneTypingInterval);
		}

		updateSlackPath();
	});

	// Update slack logo image
	$("#slack-img").change(function(e) {
		$("#slack-img-nothing-selected").hide();
		$("#slack-img-tag").css("display", "inline-flex");
		$("#slack-img-tag span").text(e.target.files[0].name);
	});

	// Update slack banner image
	$("#slack-banner-img").change(function(e) {
		$("#slack-banner-img-nothing-selected").hide();
		$("#slack-banner-img-tag").css("display", "inline-flex");
		$("#slack-banner-img-tag span").text(e.target.files[0].name);
	});

	// Validate & update slack name on load
	validateSlackName();
	updateSlackPath();		
});

function validateSlackName() {
	var field = $('#slack-name').parent().parent();
	var control = $('#slack-name').parent();
	var input = $('#slack-name');

	clearInputValidation(input);

	if (input.val().length == 0)
		return;

	$.getJSON("/api/slacks/exists/"+input.val(), function(data) {
		if (data.exists) {
			addInputErrorMsg(input, "This slack name is not available");
		} else {
			addInputSuccessMsg(input, "This slack name is available");
		}
	});
}

function updateSlackPath() {
	var slackPath = $('#slack-name').val().replace(/\s/g, '_');
	slackPath = slackPath.replace(/[\W]+/g, '');
	$("#slack-path").text("/" + slackPath.toLowerCase());
}