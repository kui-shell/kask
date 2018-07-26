
const host = "s3-api.us-geo.objectstorage.softlayer.net"; // input
const version = "1.6.1"; // intput
const inputFile = '/Users/sarahb/go/src/cloud-shell-cli/bin/sampleInput.txt'; //input

function generate(host, version, inputfile, callback) {
	const binaries = [];
	const template = {
		"plugins": [
			{
				"name": "shell-test",
				"aliases": null,
				"description": "Shell test",
				"created": "2016-01-14T00:00:00Z",
				"updated": "2018-07-05T00:00:00Z",
				"company": "IBM",
				"homepage": "https://plugins.ng.bluemix.net",
				"authors": [],
				"versions": [
				]
			}
		]
	};

	const platformMap = {
		"darwin-amd64": "osx",
		"linux-amd64": "linux64",
		"windows-amd64": "win64",
		"windows-386": "win",
		"linux-386": "linux"
	};
	const versionTemplate = {
		"version": "", //todo
		"updated": "", //todo
		"doc_url": "",
		"min_cli_version": "",
		"binaries": [], //todo
		"api_versions": null,
		"releaseNotesLink": ""
	};
	var lineReader = require('readline').createInterface({
		input: require('fs').createReadStream(inputfile)
	});

	lineReader.on('line', function (line) {
		const splitline = line.split(/[ ]+/);
		const exarray = splitline[1].split("/");
		const executable = exarray[exarray.length - 1];
		const platforms = executable.match(/cloud-shell-(.*)/);
		var platform = platforms ? platforms[1].split(".")[0] : "";
		const url = `https://${host}/shelldist/dist/${version}/${executable}`;
		binaries.push({
			"platform": platformMap[platform] || platform,
			"url": url,
			"checksum": splitline[0]
		});
	});

	lineReader.on('close', function () {
		var output = template;
		const shelltestPlugin = output.plugins.find((plugin) => plugin.name === "shell-test");
		versionTemplate.version = version;
		versionTemplate.updated = new Date();
		versionTemplate.binaries.push(binaries);
		shelltestPlugin.versions.push(versionTemplate);
		shelltestPlugin.updated = new Date();
		callback(null, output);
	});

	lineReader.on('err', function(err) {
		callback(err);
	});
}

generate(host, version, inputFile, (err, results) => {
	if (err) {
		console.log(err);
	}
	console.log(JSON.stringify(results));
});
