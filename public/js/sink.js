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
         url : "/fs/admin/sink",    
         data : {start_datetime:$("#start_datetime").val(), end_datetime:$("#end_datetime").val()},
         type : "post",  
         cache : false,  
         dataType : "json",  
         success:fsAdmin   
         }); 
    });
    $('#spnum').delegate('button[name=detail]', 'click', function() {
        $.ajax({  
         url : "/fs/adu/sink",    
         data : {start_datetime:$("#start_datetime").val(), end_datetime:$("#end_datetime").val(), spnum:$(this).val()},
         type : "post",  
         cache : false,  
         dataType : "json",  
         success:fsAdu   
         }); 
    });
});
    

function fsAdmin(json) {
    if (json.Status !== undefined) {
        $('#myModal').modal('toggle');
        $('#myModal p').text(json.Status + ":" + json.Text);
        return
    }
    var htmls = [];
    $.each(json, function(i, field){
        htmls.push("<tr>");
        htmls.push("<td>" + field.Spnum + "</td>");
        htmls.push("<td>" + field.Spname + "</td>");
        $.each(field.Data, function(key, value) {
            htmls.push("<td>" + value + "</td>");
        });
        htmls.push('<td><button class="btn btn-primary" name="detail" value="' + field.Spnum + '">明细</button></td>');                  
        htmls.push("</tr>");
    });
    $("#spnum tbody").html(htmls.join(""));
}

function fsAdu(json) {
    if (json.Status !== undefined) {
        $('#myModal').modal('toggle');
        $('#myModal p').text(json.Status + ":" + json.Text);
        return
    }
    var htmls = [];
    
    $.each(json.Consign, function(i, field){
        htmls.push("<tr>");
        htmls.push("<td>" + field.Consignid + "</td>");
        htmls.push("<td>" + field.Consignname + "</td>");        
        $.each(field.Data, function(key, value) {
            htmls.push("<td>" + value + "</td>");
        });
        htmls.push("</tr>");
    });
    $("#consign tbody").html(htmls.join(""));
    
    htmls = []
    $.each(json.Service, function(i, field){
        htmls.push("<tr>");
        htmls.push("<td>" + field.Serviceid + "</td>");
        htmls.push("<td>" + field.Servicename + "</td>");   
        $.each(field.Data, function(key, value) {
            htmls.push("<td>" + value + "</td>");
        });
        htmls.push("</tr>");
    });
    $("#serviceid tbody").html(htmls.join(""));
    
    htmls = []
    $.each(json.Province, function(i, field){
        htmls.push("<tr>");
        htmls.push("<td>" + field.Provinceid + "</td>");
        htmls.push("<td>" + field.Provincename + "</td>");          
        $.each(field.Data, function(key, value) {
            htmls.push("<td>" + value + "</td>");
        });
        htmls.push("</tr>");
    });
    $("#province tbody").html(htmls.join(""));  
}

