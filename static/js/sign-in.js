var lastSignInEmailVal = "";
var lastSignInPasswordVal = "";
var lastSignUpDisplayNameVal = "";
var lastSignUpEmailVal = "";
var lastSignUpPasswordVal = "";

var typingTimer;
var doneTypingInterval = 500; // Time in ms

$(document).ready(function() {
	// On sign in submit click then sign user in
	$("#signin-form").submit(function(e) {
		e.preventDefault();
		signIn();
	});

	// On sign up submit click then sign user in
	$("#signup-form").submit(function(e) {
		e.preventDefault();
		signUp();
	});

	// Clear validation on blur if input changed
	$("#signin-email").on("blur", function() {
		if ($(this).val() != lastSignInEmailVal) {
			clearInputValidation($(this));
		}
	});

	$("#signin-password").on("blur", function() {
		if ($(this).val() != lastSignInPasswordVal) {
			clearInputValidation($(this));
		}
	});

	$("#signup-display-name").on("blur", function() {
		if ($(this).val() != lastSignUpDisplayNameVal) {
			clearInputValidation($(this));
		}
	});

	$("#signup-email").on("blur", function() {
		if ($(this).val() != lastSignUpEmailVal) {
			clearInputValidation($(this));
		}
	});

	$("#signup-password").on("blur", function() {
		if ($(this).val() != lastSignUpPasswordVal) {
			clearInputValidation($(this));
		}
	});

	// Switching between sign in and sign up
	$("#switch-to-signup").click(switchToSignUp);
	$("#switch-to-signin").click(switchToSignIn);

	// Nav bar sign in click
	$("#signin-button").click(switchToSignIn);

	// Check if display name exists on finish typing
	$("#signup-display-name").keyup(function() {
		clearInputValidation($(this));
		clearTimeout(typingTimer);
		if ($("#signup-display-name").val()) {
			typingTimer = setTimeout(validateDisplayName, doneTypingInterval);
		}
	});
});

function validateDisplayName() {
	var input = $('#signup-display-name');

	clearInputValidation(input);

	if (input.val().length == 0)
		return;

	$.getJSON("/api/displayNames/exists/"+input.val(), function(data) {
		if (data.exists) {
			addInputErrorMsg(input, "Display name is not available");
		} else {
			addInputSuccessMsg(input, "Display name is available");
		}
	});
}

function signIn() {
	var email = $("#signin-email").val();
	var password = $("#signin-password").val();
	lastSignInEmailVal = email;
	lastSignInPasswordVal = password;

	// Clear previous validation
	clearInputValidation($("#signin-email"));
	clearInputValidation($("#signin-password"));
	
	// Validate fields
	var wasError = false;
	if (!isEmail(email)) {
		addInputErrorMsg($("#signin-email"), "Invalid email address");
		wasError = true;
	}
	if (password.length <= 2) {
		addInputErrorMsg($("#signin-password"), "Password is too short");
		wasError = true;
	}
	if (wasError) return;

	// Disabled email sign in fields
	$("#signin-email").prop("disabled", true);
	$("#signin-password").prop("disabled", true);
	$("#signin-submit").addClass("is-loading");


	// Attempt to sign in
	$.post("/do/signIn/email", {
		email: email,
		password: password,
	}, function(resp, status) {
		switch (status) {
			case "success":
				if (!resp.success) {
					var error = resp.error ? resp.error : "Something went wrong when signing in";
					alertify.error(error);
					return;
				}
				alertify.success("Signed in successfully. Reloading...");
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
	}).always(function() {
		$("#signin-email").prop("disabled", false);
		$("#signin-password").prop("disabled", false);
		$("#signin-submit").removeClass("is-loading");
	});
}

function signUp() {
	clearTimeout(typingTimer);
	
	// Retreive email sign in values
	var displayName = $("#signup-display-name").val();
	var email = $("#signup-email").val();
	var password = $("#signup-password").val();
	var retypePassword = $("#signup-retype-password").val();
	lastSignUpDisplayNameVal = displayName;
	lastSignUpEmailVal = email;
	lastSignUpPasswordVal = password;

	// Clear previous validation
	clearInputValidation($("#signup-display-name"));
	clearInputValidation($("#signup-email"));
	clearInputValidation($("#signup-password"));
	
	var wasError = false;

	// Validate display name
	if (displayName.length <= 3) {
		addInputErrorMsg($("#signup-display-name"), "Display name must be 4 or more characters in length");
		wasError = true;
	} else if (displayName.length > 32) {
		addInputErrorMsg($("#signup-display-name"), "Display name is too long");
		wasError = true;
	} else if (!/^[a-zA-Z0-9_.-]*$/.test(displayName)) {
		addInputErrorMsg($("#signup-display-name"), "Display name must contain only letters, numbers, underscores and hyphens");
		wasError = true;
	}

	// Check if email is valid
	if (!isEmail(email)) {
		addInputErrorMsg($("#signup-email"), "Invalid email address");
		wasError = true;
	}

	// Check if password is secure
	if (!passwordIsSecure(password)) {
		addInputErrorMsg($("#signup-password"), "Password must contain:<br>- 8 or more characters<br>- One or more lower case letters<br>- One or more upper case letters<br>- One or more digits");
		wasError = true;
	} else if (password.length > 64) {
		addInputErrorMsg($("#signup-password"), "Password is too long");
		wasError = true;
	}

	// Check if password's match
	if (password !== retypePassword) {
		addInputErrorMsg($("#signup-retype-password"), "Passwords do not match");
		wasError = true;
	}

	// If error, exit
	if (wasError) return;

	// Disabled sign up fields
	$("#signup-display-name").prop("disabled", true);
	$("#signup-email").prop("disabled", true);
	$("#signup-password").prop("disabled", true);
	$("#signup-submit").addClass("is-loading");


	// Attempt to sign up
	$.post("/do/signUp/email", {
		displayName: displayName,
		email: email,
		password: password,
	}, function(resp, status) {
		switch (status) {
			case "success":
				if (!resp.success) {
					var error = resp.error ? resp.error : "Something went wrong when signing up";
					alertify.error(error);
					return;
				}

				alertify.success("Successfully created account - Check your email to active your account");
				setTimeout(function() {
					alertify.log("Signing in with email " + email);
					// Attempt to sign in
					$.post("/do/signIn/email", {
						email: email,
						password: password,
					}, function(resp, status) {
						switch (status) {
							case "success":
								if (!resp.success) {
									var error = resp.error ? resp.error : "Something went wrong when signing in";
									alertify.error(error);
									return;
								}
								alertify.success("Signed in successfully. Reloading...");
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
				}, 1000);
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
	}).always(function() {
		$("#signup-display-name").prop("disabled", false);
		$("#signup-email").prop("disabled", false);
		$("#signup-password").prop("disabled", false);
		$("#signup-submit").removeClass("is-loading");
	});
}

function switchToSignUp() {
	$(".signin-container").removeClass("is-active");
	$(".signup-container").addClass("is-active");
}

function switchToSignIn() {
	$(".signup-container").removeClass("is-active");
	$(".signin-container").addClass("is-active");
}

function isEmail(str) {
	var pattern = /^(((([a-zA-Z]|\d|[!#\$%&'\*\+\-\/=\?\^_`{\|}~]|[\x{00A0}\-\x{D7FF}\x{F900}\-\x{FDCF}\x{FDF0}\-\x{FFEF}])+(\.([a-zA-Z]|\d|[!#\$%&'\*\+\-\/=\?\^_`{\|}~]|[\x{00A0}\-\x{D7FF}\x{F900}\-\x{FDCF}\x{FDF0}\-\x{FFEF}])+)*)|((\x22)((((\x20|\x09)*(\x0d\x0a))?(\x20|\x09)+)?(([\x01-\x08\x0b\x0c\x0e-\x1f\x7f]|\x21|[\x23-\x5b]|[\x5d-\x7e]|[\x{00A0}\-\x{D7FF}\x{F900}\-\x{FDCF}\x{FDF0}\-\x{FFEF}])|(\([\x01-\x09\x0b\x0c\x0d-\x7f]|[\x{00A0}\-\x{D7FF}\x{F900}\-\x{FDCF}\x{FDF0}\-\x{FFEF}]))))*(((\x20|\x09)*(\x0d\x0a))?(\x20|\x09)+)?(\x22)))@((([a-zA-Z]|\d|[\x{00A0}\-\x{D7FF}\x{F900}\-\x{FDCF}\x{FDF0}\-\x{FFEF}])|(([a-zA-Z]|\d|[\x{00A0}\-\x{D7FF}\x{F900}\-\x{FDCF}\x{FDF0}\-\x{FFEF}])([a-zA-Z]|\d|-|\.|_|~|[\x{00A0}\-\x{D7FF}\x{F900}\-\x{FDCF}\x{FDF0}\-\x{FFEF}])*([a-zA-Z]|\d|[\x{00A0}\-\x{D7FF}\x{F900}\-\x{FDCF}\x{FDF0}\-\x{FFEF}])))\.)+(([a-zA-Z]|[\x{00A0}\-\x{D7FF}\x{F900}\-\x{FDCF}\x{FDF0}\-\x{FFEF}])|(([a-zA-Z]|[\x{00A0}\-\x{D7FF}\x{F900}\-\x{FDCF}\x{FDF0}\-\x{FFEF}])([a-zA-Z]|\d|-|\.|_|~|[\x{00A0}\-\x{D7FF}\x{F900}\-\x{FDCF}\x{FDF0}\-\x{FFEF}])*([a-zA-Z]|[\x{00A0}\-\x{D7FF}\x{F900}\-\x{FDCF}\x{FDF0}\-\x{FFEF}])))\.?$/g
	return pattern.test(str);
}

function passwordIsSecure(str) {
	var pattern = /^(?=.*?[A-Z])(?=(.*[a-z]){1,})(?=(.*[\d]){1,})(?=(.*[\W]?))(?!.*\s).{8,}$/g
	return pattern.test(str);
}

// Google sign in
initGoogleSignIn();
function initGoogleSignIn() {
	$(".google-signin").click(function() {
		auth2.grantOfflineAccess().then(function(codeData) {
			if (!codeData) {
				alertify.error("Something went wrong");
				return;
			}

			// var basicProfile = googleUser.getBasicProfile();
			$.post("/do/signIn/google", 
			// {
			// 	id: basicProfile.getId(),
			// 	email: basicProfile.getEmail(),
			// 	firstName: basicProfile.getGivenName(),
			// 	lastName: basicProfile.getFamilyName(),
			// 	avatarURL: basicProfile.getImageUrl()
			// }
			codeData, function(resp, status) {
				switch (status) {
					case "success":
						if (!resp.success) {
							var error = resp.error ? resp.error : "Something went wrong when signing in";
							alertify.error(error);
							return;
						}
						alertify.success("Signed in successfully. Reloading...");
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

			$(".signin-container").removeClass("is-active");
		}, function(error) {
			console.log(error);
			alertify.error("Something went wrong");

			$(".signin-container").removeClass("is-active");
		});	
	});
}

// Twitter sign in
$(".twitter-signin").click(function() {
	window.location = "/do/signIn/twitter";
});

// Github sign in
$(".github-signin").click(function() {
	window.location = "https://github.com/login/oauth/authorize?scope=user:email&client_id=0592223874a6c4874d48";
});