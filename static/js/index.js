$(document).ready(function () {

    var job_list = $("#job-list");

    job_list.on("click", ".edit-job", function (event) {
        $('#edit-name').val($(this).parents('tr').children('.job-name').text());
        $('#edit-cmd').val($(this).parents('tr').children('.job-cmd').text());
        $('#edit-cronexpr').val($(this).parents('tr').children('.job-cron').text());
        var modal = $("#edit-modal");
        modal.modal("show");
    });

    $("#save-job").on('click', function () {
        var jobInfo = {
            name: $('#edit-name').val(),
            command: $('#edit-cmd').val(),
            cron_expr: $('#edit-cronexpr').val()
        };
        $.ajax({
            url: '/job/save',
            type: 'post',
            dataType: 'json',
            data: {job: JSON.stringify(jobInfo)},
            complete: function () {
                window.location.reload();
            }
        });
    });

    $("#new-job").on('click', function () {

        $('#edit-name').val("");
        $('#edit-cmd').val("");
        $('#edit-cronexpr').val("");
        var modal = $("#edit-modal");
        modal.modal("show");
    });

    job_list.on("click", ".del-job", function (event) {

        var jobName = $(this).parents("tr").children(".job-name").text();
        console.log("delete: " + jobName);
        $.ajax({
            url: "/job/del",
            type: "post",
            dataType: 'json',
            data: {name: jobName},
            complete: function () {
                window.location.reload();
            }
        })

    });

    job_list.on("click", ".kill-job", function (event) {
        var jobName = $(this).parents("tr").children(".job-name").text();
        console.log("delete: " + jobName);
        $.ajax({
            url: "/job/kill",
            type: "post",
            dataType: 'json',
            data: {name: jobName},
            complete: function () {
                // window.location.reload();
            }
        })
    });

    function rebuild_list() {
        $.ajax({
            url: "/job/list",
            dataType: "json",
            success: function (resp) {
                if (resp.errno !== 0) {
                    return;
                }
                $("#job-list tbody").empty();
                var job_list = resp.data;
                for (var i = 0; i < job_list.length; ++i) {
                    var job = job_list[i];
                    var tr = $("<tr>");
                    tr.append($('<td class="job-name">').html(job.name));
                    tr.append($('<td class="job-cmd">').html(job.command));
                    tr.append($('<td class="job-cron">').html(job.cron_expr));

                    var toolbar = $('<div class="btn-toolbar">').
                    append('<button class="btn btn-info edit-job">编辑</button>').
                    append('<button class="btn btn-danger del-job">删除</button>').
                    append('<button class="btn btn-warning kill-job">杀死</button>');

                    tr.append($('<td>').append(toolbar));
                    $("#job-list tbody").append(tr);
                }
            }
        })

    }

    rebuild_list()
});