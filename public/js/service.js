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
         url : "/ur",    
         data : {start_datetime:$("#start_datetime").val(), end_datetime:$("#end_datetime").val(), terminal:$("#terminal").val()},
         type : "post",  
         cache : false,  
         dataType : "json",  
         success:ur   
         }); 
    });
});
    


function ur(json) {
    if (json.Status !== undefined) {
        $('#myModal').modal('toggle');
        $('#myModal p').text(json.Status + ":" + json.Text);
        return
    }
    var htmls = [];
    var moNum = 0;
    if (json.Mos !== null) {
    $.each(json.Mos, function(i, field){        
        htmls.push("<tr>");
        htmls.push("<td>" + field.spinfo.spnum + "</td>");
        htmls.push("<td>" + field.spinfo.spname + "</td>");        
//        htmls.push("<td>" + field.spconsign.consignid + "</td>");        
        htmls.push("<td>" + field.spconsign.consignname + "</td>");        
        htmls.push("<td>" + field.spservice.serviceword + "</td>");        
        htmls.push("<td>" + field.spservice.servicename + "</td>");        
        htmls.push("<td>" + field.spservice.servicefee + "</td>");        
        htmls.push("<td>" + field.spmsisdn.provincename+ "</td>"); 
        htmls.push("<td>" + field.spmsisdn.cityname+ "</td>"); 
        htmls.push("<td>" + field.spservicerule.terminal+ "</td>"); 
        htmls.push("<td>" + field.spservicerule.statusid+ "</td>");
        htmls.push("<td>" + field.spservicerule.linkid+ "</td>");
//        htmls.push("<td>" + field.spservicerule.expendtime+ "</td>");
        htmls.push("<td>" + field.spservicerule.timeline+ "</td>");  
        htmls.push("</tr>");       
    });
        moNum = json.Mos.length;
    }
    $("#moNum").html('<span class="badge badge-success">' + moNum + '</span>')            
    $("#mo tbody").html(htmls.join(""));

    var htmls = [];
    var mtNum = 0;
    var smtNum = 0;
    var fee = 0;
    if (json.Mts !== null) {
    $.each(json.Mts, function(i, field){
        htmls.push("<tr>");
        htmls.push("<td>" + field.spinfo.spnum + "</td>");
        htmls.push("<td>" + field.spinfo.spname + "</td>");        
//        htmls.push("<td>" + field.spconsign.consignid + "</td>");        
        htmls.push("<td>" + field.spconsign.consignname + "</td>");        
        htmls.push("<td>" + field.spservice.serviceword + "</td>");        
        htmls.push("<td>" + field.spservice.servicename + "</td>");        
        htmls.push("<td>" + field.spservice.servicefee + "</td>");        
        htmls.push("<td>" + field.spmsisdn.provincename+ "</td>"); 
        htmls.push("<td>" + field.spmsisdn.cityname+ "</td>"); 
        htmls.push("<td>" + field.spservicerule.terminal+ "</td>"); 
        htmls.push("<td>" + field.spservicerule.statusid+ "</td>");
        htmls.push("<td>" + field.spservicerule.linkid+ "</td>");
        htmls.push("<td>" + field.spservicerule.expendtime+ "</td>");
        htmls.push("<td>" + field.spservicerule.timeline+ "</td>");                        
        htmls.push("</tr>");
        if (field.spservicerule.statusid === "200") {
            smtNum += 1
            fee += field.spservice.servicefee
        }
    });
      mtNum = json.Mts.length; 
    }
    $("#mtNum").html('条数<span class="badge badge-success">' + mtNum + '</span>' + '成功条数<span class="badge badge-success">' + smtNum + '</span>' + '金额/(分)<span class="badge badge-important">' + fee + '</span>')     
    $("#mt tbody").html(htmls.join(""));

  
    
}
