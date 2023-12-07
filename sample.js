// 1. init ... this section is called per VU
import otlpext from 'k6/x/otlpext';
import {sleep} from 'k6';

let localCounter = 0;


// 2. setup ... this function is called only once
export function setup(){
	console.log("*** setup ***");
	let serviceName = "hoge";
	if (__ENV.SERVICENAME){
		serviceName = __ENV.SERVICENAME;
	}
	otlpext.initialize("http://localhost:4317", serviceName);
}

// 3. default ... this function is called many timed per iteration per VU
export default function(){
	let traceId = otlpext.sendTrace("hoge");
	// this function is called per -u (local counter is per user)
	localCounter++;
	console.log(`${traceId} sent. otlpext counter:${otlpext.counter} local counter:${localCounter}`);
	// sleep(1);
}

// 4. teardown ... this function is called only once
export function teardown(){
	console.log("*** teardown ***");
	otlpext.shutdown();
}
