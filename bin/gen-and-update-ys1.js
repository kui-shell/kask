const fs = require("fs");
const request = require("request");

const args = process.argv;
if (args.length < 4) {
	console.error(`USAGE: node gen-and-update-ys1.js "1.6.1" "userName" "password"`);
	process.exit(1);
}

const PLUGIN_NAME = "function-composer";
const versionArg = args[2];
const userName = args[3];
const password = args[4];
const inputFile = "./out/checksums";

const platformConfig = {
	"windows-386.exe": "Checksum_Win_X86",
	"windows-amd64.exe": "Checksum_Win_X64",
	"linux-386": "Checksum_Linux_X86",
	"linux-amd64": "Checksum_Linux_X64",
	"darwin-amd64": "Checksum_MacOS"
};

function hasExistingVersion(currentVersion, callback) {
	request({
		url: "https://plugins.stage1.ng.bluemix.net/bx/list/" + PLUGIN_NAME,
		json: true
	}, (err, response, body) => {
		if (err) {
			console.error(err);
			console.error(`Couldn't get existing versions!`);
			process.exit(1);
		}
		const existing = body.findIndex((version) => version.version === currentVersion) > -1;
		console.log(`Version ${currentVersion} exists: ` + existing);
		callback(null, existing);
	});
}

function generate(version, inputFile, userName, password) {
	const formData = {};
	const params = [];
	params.push({
		"name": "Plugin_Name",
		"value": PLUGIN_NAME
	});
	params.push({
		"name": "Version",
		"value": version
	});
	const lines = fs.readFileSync(inputFile, "utf-8").split('\n').filter(Boolean);
	lines.forEach((line) => {
		const regExp = new RegExp(`(\\S*)\\s+(.*\/(${PLUGIN_NAME}-(.*)-(.*)))$`);
		const match = line.match(regExp);
		const fullFileName = match[2];
		const fileName = match[3];
		const platform = `${match[4]}-${match[5]}`;
		const checksum = match[1];
		const shortFileName = platform.replace(".exe", "");
		params.push({
			"name": platform,
			"file": shortFileName
		});
		params.push({
			"name": platformConfig[platform],
			"value": checksum
		});
		formData[shortFileName] = fs.createReadStream(fullFileName);
	});
	formData["json"] = JSON.stringify({
		parameter: params
	});

	hasExistingVersion(versionArg, (err, exists) => {
		const jobName = exists ? "Refresh-Plugin-Version-on-YS1" : "Publish%20Plugin%20to%20YS1";
		request.post({
			url: `https://wcp-cloud-foundry-jenkins.swg-devops.com/job/${jobName}/build`,
			formData: formData,
			auth: {
				user: userName,
				pass: password
			}
		}, (err, response, body) => {
			if (err) {
				console.error('upload failed:', err);
				process.exit(1);
			}
			if (response.statusCode < 200 || response.statusCode > 299) {
				console.error('upload failed:', response.statusCode + " - " + body);
				process.exit(1);
			}
			console.log('Upload successful!  Server responded with:', body);
		});
	});
}

generate(versionArg, inputFile, userName, password);
