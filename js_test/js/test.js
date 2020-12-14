var ingress = "http://10.64.140.43";
var loginService = ingress + "/greeter/auth";
var greeterService = ingress + "/greeter/messages";

var loginUser = "tobo";
var loginPass = "obot";

$(document).ready(function() {
	//
	// Login
	//
	var loginServiceUsers = loginService + "/users&callback=?";
	var loginServiceLogin = loginService + "/login";

	var creds = {
        "user_name": loginUser,
        "user_password": loginPass
	};

	$.ajax({
        url: loginServiceUsers,
        method: "POST",
        contentType: "application/json",
        data: JSON.stringify(creds)
    }).fail(function(resp) {
        console.log(resp.status)
    }).done(function(resp) {
        console.log(resp)
    });
})