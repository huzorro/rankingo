$("#start_datetime").datetimepicker({format: 'yyyy-mm-dd',
                                     language: 'zh-CN',
                                     todayBtn: 'linked',
                                     minView: 4,
                                     autoclose: true,
                                     pickerPosition: "bottom-left"
                                    }).on("changeDate", function(ev) {
                                        console.log(ev.date.valueOf())
                                    });

$("#end_datetime").datetimepicker({format: 'yyyy-mm-dd', 
                                   language: 'zh-CN', 
                                   todayBtn: 'linked', 
                                   minView: 4, 
                                   autoclose: true, 
                                   pickerPosition: "bottom-right"}).on("changeDate", function(ev) {
                                        console.log($("#end_datetime").val());
                                   });


$(function() {
    $("#search").click(function() {
        $.ajax({  
         url : "/user",    
         data : {start_datetime:$("#start_datetime").val(), end_datetime:$("#end_datetime").val()},
         type : "post",  
         cache : false,  
         dataType : "json",  
         success:user   
         }); 
    });
});
    


function user(json) {
    if (json.Status !== undefined) {
        $('#myModal').modal('toggle');
        $('#myModal p').text(json.Status + ":" + json.Text);
        return
    }
    var htmls = [];
    if (json.Consign !== null) {
    $.each(json.Consign, function(i, field){
        htmls.push("<tr>");
        htmls.push("<td>" + field.Consignid + "</td>");
        htmls.push("<td>" + field.Consignname + "</td>");        
        $.each(field.Data, function(key, value) {
            htmls.push("<td>" + value + "</td>");
        });
        htmls.push("</tr>");
    });
    }
    $("#consign tbody").html(htmls.join(""));
    
    htmls = []
    if (json.Service !== null) {
    $.each(json.Service, function(i, field){
        htmls.push("<tr>");
        htmls.push("<td>" + field.Serviceid + "</td>");
        htmls.push("<td>" + field.Servicename + "</td>");   
        $.each(field.Data, function(key, value) {
            htmls.push("<td>" + value + "</td>");
        });
        htmls.push("</tr>");
    });
    }
    $("#serviceid tbody").html(htmls.join(""));
    
    htmls = []
    if (json.Province !== null) {
    $.each(json.Province, function(i, field){
        htmls.push("<tr>");
        htmls.push("<td>" + field.Provinceid + "</td>");
        htmls.push("<td>" + field.Provincename + "</td>");          
        $.each(field.Data, function(key, value) {
            htmls.push("<td>" + value + "</td>");
        });
        htmls.push("</tr>");
    });
    }
    $("#province tbody").html(htmls.join(""));  
}
