$(document).ready(function() {
	
	$("#login").click(function() {
		console.log("in login");
		el = document.getElementById("loginBox");
		el.style.visibility = (el.style.visibility == "visible") ? "hidden" : "visible";
	});
	
	$("#close").click(function() {
		console.log("in close");
		el = document.getElementById("loginBox");
		el.style.visibility = (el.style.visibility == "visible") ? "hidden" : "visible";
	});
	
	$("#loginForm").submit(function() {
		//document.getElementById('errorLabel').innerHTML('no error');
		var email =$("input#email_label").val();
		var password=$("input#password_label").val();
		$.ajax({
			type:"POST",
			url:"http://1b7db24b.ngrok.com/login",
			data:{"email" : email, "pass" : password},
			success: function(html) {
				if (html=='Yes') {
					console.log("**********************")
					setTimeout('go_to_userPage()', 500);
				} else {
					document.getElementById("errorLabel").innerHTML="wrong username or password";
				}
			}
		});
		
		return false;
	});
	
	
});

function go_to_userPage() {
	window.location="http://1b7db24b.ngrok.com/auth"
}

function registerBox() {
	el = document.getElementById("registerBox");
	el.style.visibility = (el.style.visibility == "visible") ? "hidden" : "visible";
}
