
$(function() {
    $("#addKey").click(function() {
        $.ajax({
         url : "/key/add",
         data : {Keyword:$("input[name=Keyword]").val(), Destlink:$("input[name=Destlink]").val(),
                 KeyCity:$("input[name=KeyCity]").val(), KeyProvince:$("input[name=KeyProvince]").val()},
         type : "post",
         cache : false,
         dataType : "json",
         success: keyadd
         });
    });
//    $("#my-tab-content").delegate("button[name=operate]", "click", function() {
//         console.log($(this).val());
//        $.ajax({
//         url : "/key/one",
//         data : {Id:$(this).val()},
//         type : "post",
//         cache : false,
//         dataType : "json",
//         success: keyone
//         });
//    });
    $("button[name=operate]").click(function() {
         console.log($(this).val() + "Abc");
        $.ajax({
         url : "/key/one",
         data : {Id:$(this).val()},
         type : "post",
         cache : false,
         dataType : "json",
         success: keyone
         });
    });
    $("#updateKey").click(function() {
        $.ajax({
         url : "/key/update",
         data : {Id:$("#Id").val(), Owner:$("#Owner").val(), Status:$("input[name=Status]:checked").val()},
         type : "post",
         cache : false,
         dataType : "json",
         success: keyupdate
         });
    });
    $("#infoModal .btn").click(function(){
        location.reload();
    });
});


function keyadd(json) {
    if (json.status !== undefined) {
        $('#infoModal').modal('toggle');
        $('#infoModal p').text(json.text);
//        location.reload();
        return
    }
}


function keyupdate(json) {
    if (json.status !== undefined) {
        $('#infoModal').modal('toggle');
        $('#infoModal p').text(json.text);
//        location.reload();
        return
    }
}
function keyone(json) {
        console.log("abc");
        $('#updateKeyModal').modal('toggle');
        if (json.status !== undefined) {
            $("#keyStatus").html(json.text);
        } else {
            var htmls = [];
            $("#Id").val(json.keyMsg.id);
            $("#Owner").val(json.keyMsg.owner);
            if (!json.cancel) {
                htmls.push('<input type="radio" name="Status" value=0 checked>优化<input type="radio" name="Status" value=1>停止');
            } else {
                htmls.push('<input type="radio" name="Status" value=0>优化<input type="radio" name="Status" value=1 checked>停止');
            }
            $("#keyStatus").html(htmls.join(""));
        }
}
