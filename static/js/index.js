$(document).ready(function () {

    function timeFormat(millsecond) {
        function paddingNum(num, n) {
            var len = num.toString().length;
            while (len < n) {
                num = '0' + num;
                len++;
            }
            return num
        }

        var date = new Date(millsecond);
        var year = date.getFullYear();
        var month = paddingNum(date.getMonth() + 1, 2);
        var day = paddingNum(date.getDate(), 2);
        var hour = paddingNum(date.getHours(), 2);
        var min = paddingNum(date.getMinutes(), 2);
        var second = paddingNum(date.getSeconds(), 2);
        var mill = paddingNum(date.getMilliseconds(), 3);

        return year + "-" + month + "-" + day + " " + hour + ":" + min + ":" + second + "." + mill;
    }

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

    job_list.on("click", ".log-job", function (event) {
        $("#log-list tbody").empty();
        var job_name = $(this).parents('tr').children(".job-name").text();
        $("#log-title").text(job_name + "的日志")
        console.log("日志显示：" + job_name);

        $.ajax({
            url: "/job/log",
            type: "post",
            dataType: "json",
            data: {name: job_name, skip: 0, limit: 20},
            success: function (resp) {
                if (resp.errno != 0) {
                    return;
                }

                var logList = resp.data;
                for (var i = 0; i < logList.length; ++i) {
                    var log = logList[i];
                    var tr = $('<tr>');
                    tr.append($('<td>').html(log.command));
                    tr.append($('<td>').html(log.err));
                    tr.append($('<td>').html(log.output));
                    tr.append($('<td>').html(timeFormat(log.plan_time)));
                    tr.append($('<td>').html(timeFormat(log.schedule_time)));
                    tr.append($('<td>').html(timeFormat(log.exec_start_time)));
                    tr.append($('<td>').html(timeFormat(log.exec_end_time)));
                    console.log(tr);
                    $("#log-list tbody").append(tr);
                }
            }
        });

        $("#log-modal").modal("show");

    });

    $("#list-workers").on("click", function () {
       $.ajax({
           url: "/worker/list",
           dataType: 'json',
           success: function (resp) {
               if (resp.errno != 0) {
                   return;
               }
               var workers = resp.data;
               for (var i = 0; i < workers.length; ++i) {
                   var id = workers[i].id;
                   var tr = $('<tr>');
                   tr.append($('<td>').html(id));
                   $('#worker-list tbody').append(tr);
               }
           }
       });

        $('#worker-modal').modal('show');
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
                    append('<button class="btn btn-warning kill-job">杀死</button>').
                    append('<button class="btn btn-success log-job">日志</button>');

                    tr.append($('<td>').append(toolbar));
                    $("#job-list tbody").append(tr);
                }
            }
        })

    }

    rebuild_list()
});