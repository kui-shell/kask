const fs = require("fs");
const request = require("request");

const args = process.argv;
if (args.length < 4) {
	console.error(`USAGE: node generate-plugin-json.js "https://s3-api.us-geo.objectstorage.softlayer.net/shelldist/dist" "1.6.1" "api-key"`);
	process.exit(1);
}

const hostArg = args[2];
const versionArg = args[3];
const apiKey = args[4];
const inputFile = "./out/checksums";
const outputFile = "./out/plugins.json";
let accessToken;

const platformMap = {
	"darwin-amd64": "osx",
	"linux-amd64": "linux64",
	"windows-amd64": "win64",
	"windows-386": "win",
	"linux-386": "linux"
};

function generate(host, version, inputFile) {
	host = host.endsWith("/") ? host.substring(0, host.length -1) : host;
	const json = {
		"plugins": [
			{
				"name": "cloud-shell",
				"aliases": null,
				"description": "Cloud shell",
				"created": "2016-01-14T00:00:00Z",
				"updated": "2018-07-05T00:00:00Z",
				"company": "IBM",
				"homepage": "https://plugins.ng.bluemix.net",
				"authors": [],
				"versions": [createVersion(host, version, inputFile)]
			}
		]
	};
	return json;
}

function createVersion(host, version, inputFile) {
	return {
		"version": version,
		"updated": new Date(),
		"doc_url": "",
		"min_cli_version": "",
		"api_versions": null,
		"releaseNotesLink": "",
		"binaries": createBinaries(host, version, inputFile)
	};
}

function createBinaries(host, version, inputFile) {
	const lines = fs.readFileSync(inputFile, "utf-8").split('\n').filter(Boolean);
	return lines.map((line) => {
		const splitline = line.match(/(^\S*).*\/cloud-shell-(.*)-([^\.]*)(\.exe)*$/);
		const platform = `${splitline[2]}-${splitline[3]}`;
		const response = {
			"platform": platformMap[platform] || platform,
			"url": `${host}/${version}/cloud-shell-${platform}-${version}${splitline[4] || ""}`,
			"checksum": splitline[1]
		};
	    uploadExecutable(response.url, fs.readFileSync(`../out/cloud-shell-${platform}${splitline[4] || ""}`));
		return response;
	});
}

function getIAMToken(callback) {
	request.post({
		url: "https://iam.bluemix.net/oidc/token",
		headers: {
			"Accept": "application/json",
			"Content-type": "application/x-www-form-urlencoded"
		},
		form: {
			apikey: apiKey,
			response_type: "cloud_iam",
			grant_type: "urn:ibm:params:oauth:grant-type:apikey"
		}
	}, callback);
}


function uploadExecutable(url, binary) {
	if (binary) {
		request.put({
			url: url,
			headers: {
				"x-amz-acl": "public-read",
				"Authorization": `Bearer ${accessToken}`,
				"Content-Type": "application/octet-stream",
			},
			body: binary,
			encoding: null
		}, (err, response) => {
			if (err || response.statusCode != 200) {
				console.log("failed to upload " + url);
				console.log(err || response.statusCode);
				return;
			}
			console.log(response.statusCode + " on upload of " + url);
		});
	}
}


// Main
getIAMToken((err, response) => {
	if (err || response.statusCode != 200) {
		console.log("upload failed due to no iam-token.");
		return;
	}
	let body;
	try {
		body = JSON.parse(response.body);
	} catch(e) {
		console.log("error parsing iam-token response");
		return;
	}
	accessToken = body.access_token;
	const json = generate(hostArg, versionArg, inputFile);
	fs.writeFileSync(outputFile, JSON.stringify(json, null, 2));
});
