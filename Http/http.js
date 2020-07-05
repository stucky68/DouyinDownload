var http = require("http");
const ies = require('./public');

var onRequest = function(request,response){
    let datas = '';
	request.on('data',function(data)
	{
		datas += data;
	})
	
	request.on('end',function()
	{
		datas1 = JSON.parse(datas);
		var signature = ies.sign(datas1.uid, base64ToString(datas1.tac), datas1.ua);
		response.end(JSON.stringify(commonResponse.success(signature)));
	})
}

/**
 * 通用响应体
 */
var commonResponse = {
    success: function (result) {
        return {
            status: 0,
            message: 'success',
            result: result
        }
    },
    error: function (message) {
        return {
            status: -1,
            message: message
        }
    }
};

function base64ToString(base) {
	var b = new Buffer.from(base, 'base64')
	return b.toString();
}

var server = http.createServer(onRequest);

//最后让服务器监听一个端口
server.listen(3000,"127.0.0.1");//还可以加第二个参数 127.0.0.1代表的是本地

console.log("server started on localhost port 3000");//加一个服务器启动起来的提示