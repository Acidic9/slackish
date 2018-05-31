$(document).ready(function() {
	// Show/hide nav menu's on hamburger click
	$(".nav-toggle").click(function() {
		var isActive = $(".nav-menu").data("isActive");
		if (isActive === true) {
			$(".nav-menu").removeClass("is-active").data("isActive", false);
		} else {
			$(".nav-menu").addClass("is-active").data("isActive", true);
		}
	});

	// Nav bar drop down show/hide
	$(".open-drop-down-menu").click(function(e) {
		e.stopPropagation();

		if ($(".drop-down-menu").data("isVisible")) {
			$(".drop-down-menu").fadeOut("fast");
			$(".drop-down-menu").data("isVisible", false);
		} else {
			$(".drop-down-menu").fadeIn("fast");
			$(".drop-down-menu").data("isVisible", true);
		}
	});
	
	$('.drop-down-menu').click(function(e){
		e.stopPropagation();
	});

	$(window).click(function() {
		$(".drop-down-menu").fadeOut();
	});

	// Nav bar sign out click
	$(".signout-button").click(signOut);
});