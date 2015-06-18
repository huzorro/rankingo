
$(function() {
    $("#addUser").click(function() {
        $.ajax({  
         url : "/user/add",    
         data : {userName:$("input[name=userName]").val(),password:$("input[name=password]").val()},
         type : "post",  
         cache : false,  
         dataType : "json",  
         success: commonInfo  
         });         
    });
	  	
    $("#payment").click(function() {
        $.ajax({  
         url : "/pay",    
         data : {uid:$("#id").val(), balance:$("#balance").val(),remark:$("input[name=remark]:checked").val()},
         type : "post",  
         cache : false,  
         dataType : "json",  
         success: commonInfo  
         });         
    });    

    $("button[name=payment]").click(function() {
         console.log($(this).val() + "Abc");
		$('#paymentUserModal').modal('toggle'); 
		$("#id").val($(this).val());		      
    });
	
    $("button[name=viewUser]").click(function() {
         console.log($(this).val() + "Abc");
        $.ajax({  
         url : "/user/view",    
         data : {id:$(this).val()},
         type : "post",  
         cache : false,  
         dataType : "json",  
         success: userview   
         });       
    });
		  
    $("#updateUser").click(function() {
        $.ajax({  
         url : "/user/edit",    
         data : {id:$("#id").val(), userName:$("#userName").val(), password:$("#password").val()},                 
         type : "post",  
         cache : false,  
         dataType : "json",  
         success: commonInfo  
         });         
    });
    $("#infoModal .btn").click(function(){
        location.reload();
    });
});



function commonInfo(json) {
    if (json.status !== undefined) {
        $('#infoModal').modal('toggle');
        $('#infoModal p').text(json.text);
        return
    }    
}
function userview(json) {
        console.log("abc");
        $('#viewUserModal').modal('toggle');    
        if (json.status !== undefined) {
            commonInfo(json);
        } else {
			$("#id").val(json.Id);
            $("#userName").val(json.UserName);
            $("#password").val(json.Password);
        }
}
