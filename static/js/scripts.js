$(document).ready(function() {
	alertify.maxLogItems(8);

	// Close button on modal hides all modals
	$(".modal-close").click(function() {
		$(".modal").removeClass("is-active");
	});
});

// Clears all validation from an input
function clearInputValidation(input) {
	var field = input.parent().parent();
	var control = input.parent();

	control.removeClass("has-icon has-icon-right");
	control.children("span.icon").remove();
	input.removeClass("is-success is-danger");
	field.children("p.help").remove();
}

// Adds an error message to an input
function addInputErrorMsg(input, msg) {
	var field = input.parent().parent();
	var control = input.parent();

	field.append('<p class="help is-danger">'+msg+'</p>');

	control.addClass("has-icon has-icon-right");
	control.append('<span class="icon is-small"><i class="fa fa-warning"></i></span>');
	input.addClass("is-danger");
}

// Adds a success message to an input
function addInputSuccessMsg(input, msg) {
	var field = input.parent().parent();
	var control = input.parent();

	field.append('<p class="help is-success">'+msg+'</p>');

	control.addClass("has-icon has-icon-right");
	control.append('<span class="icon is-small"><i class="fa fa-check"></i></span>');
	input.addClass("is-success");
}

// Sign out
function signOut() {
	auth2.signOut();

	// Attempt to sign out
	$.post("/do/signOut", function(resp, status) {
		switch (status) {
			case "success":
				if (!resp.success) {
					var error = resp.error ? resp.error : "Something went wrong when signing out";
					alertify.error(error);
					return;
				}
				alertify.success("Signed out successfully. Reloading...");
				setTimeout(function() { location.reload(); }, 1000);
				break;

			case "timeout":
				alertify.error("The server didn't respond");
				break;
		
			default:
				alertify.error("Something went wrong");
				break;
		}
	}, "json").fail(function() {
		alertify.error("Something went wrong");
	});
}

// Google sign in/out
var googleUser = {};
gapi.load('auth2', function() {
	// Retrieve the singleton for the GoogleAuth library and set up the client.
	auth2 = gapi.auth2.init({
		// client_id: '83143033304-3uf2bhg6tv8gcort222d2d8dv19a3ppv.apps.googleusercontent.com',
		//cookiepolicy: 'single_host_origin',
		// Request scopes in addition to 'profile' and 'email'
		//scope: 'additional_scope'
	});
});