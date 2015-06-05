$(function() {
    $("#login").click(function() {
        $.ajax({  
         url : "/login/check",    
         data : {username:$("#username").val(), password:$("#password").val()},
         type : "post",  
         cache : false,  
         dataType : "json",  
         success:callback   
         }); 
    });  
});
    

function callback(json) {
    if (json.status !== "200") {
        $('#myModal').modal('toggle');
        $('#myModal p').text(json.status + ":" + json.text);
        console.log(json);
        return
    }
     window.location.assign("/key/show");
}

