package gypsum

const simpleHtmlPage = `<!DOCTYPE html>
<html>

<head>
    <script src="https://cdn.jsdelivr.net/npm/jquery@3.5.1/dist/jquery.min.js"></script>
    <script>
        function getForm() {
            obj = {};
            obj.group_id = $("#group_id")[0].valueAsNumber;
            obj.user_id = $("#user_id")[0].valueAsNumber;
            obj.matcher_type = $("#matcher_type")[0].valueAsNumber;
            obj.patterns = [$("#pattern")[0].value];
            obj.response = $("#response")[0].value;
            obj.priority = $("#priority")[0].valueAsNumber;
            obj.block = $("#block")[0].checked;
            return obj;
        }
        $(document).ready(function () {
            alert("此页面仅供测试，未作输入验证，任何非法输入都可能导致程序崩溃");
            $("#add_rule").click(function () {
                $.post({
                    url: "/api/v1/rules",
                    data: JSON.stringify(getForm()),
                    contentType: 'application/json',
                }).done(function (data, status) {
                    alert(JSON.stringify(data));
                })
            });
            $("#delete_rule").click(function () {
                let rule_id = $("#rule_id")[0].valueAsNumber;
				$.ajax({
					url: "/api/v1/rules/" + rule_id,
					type: "DELETE",
					success: function(result) {
						alert(JSON.stringify(result));
					}
				});
            });
        });
    </script>
</head>

<body>
    <h1>测试主页</h1>
    <div>
        <a href="/api/v1/rules" target="_blank">查看所有规则</a>
    </div>
    <br /><br />
    <div>
        <form>
            群号: <input id="group_id" type="number">（全部则填0）<br />
            QQ号: <input id="user_id" type="number">（全部则填0）<br />
            种类:
            <select id="matcher_type">
                <option value="0">完全匹配</option>
                <option value="1">关键词匹配</option>
                <option value="2">前缀匹配</option>
                <option value="3">后缀匹配</option>
                <option value="4">命令匹配</option>
                <option value="5">正则匹配</option>
            </select><br />
            匹配: <input id="pattern"><br />
            回复: <textarea id="response" cols="60" rows="5"></textarea><br />
            优先级: <input id="priority" type="number"><br />
            阻止后续: <input id="block" type="checkbox"><br />
        </form>
        <button id="add_rule">新增规则</button>
    </div>
    <br /><br />
    <div>
    规则号: <input id="rule_id" type="number">
    <button id="delete_rule">删除规则</button>
    </div>
</body>

</html>`
