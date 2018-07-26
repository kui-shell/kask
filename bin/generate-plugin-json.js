const fs = require("fs");

const args = process.argv;
if (args.length < 4) {
	console.error(`USAGE: node generate-plugin-json.js "https://s3-api.us-geo.objectstorage.softlayer.net/shelldist/dist" "1.6.1"`);
	process.exit(1);
}

const hostArg = args[2];
const versionArg = args[3];
const inputFile = "./out/checksums";
const outputFile = "./out/plugins.json";

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
		const splitline = line.split(/[ ]+/);
		const exarray = splitline[1].split("/");
		const executable = exarray[exarray.length - 1];
		const platforms = executable.match(/cloud-shell-(.*)/);
		const platform = platforms ? platforms[1].split(".")[0] : "";
		const url = `${host}/${version}/${executable}`;
		return {
			"platform": platformMap[platform] || platform,
			"url": url,
			"checksum": splitline[0]
		};
	});
}

const json = generate(hostArg, versionArg, inputFile);
fs.writeFileSync(outputFile, JSON.stringify(json, null, 2));