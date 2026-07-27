import fs from 'node:fs';
import path from 'node:path';

const root = process.cwd();
const read = (relative) => fs.readFileSync(path.join(root, relative), 'utf8');
const assertIncludes = (source, expected, label) => {
    if (!source.includes(expected)) {
        throw new Error(`${label}: missing ${expected}`);
    }
};

const logicalServiceApi = read('src/services/logical_service.ts');
const groupApi = read('src/services/config_group.ts');
const fileApi = read('src/services/config_files.ts');
const serviceDetail = read('src/pages/Discovery/Services/Instance/ServiceDetail.tsx');
const groupFiles = read('src/pages/Configuration/Group/Files/index.tsx');
const fileView = read('src/pages/Configuration/Group/Files/FileView.tsx');

assertIncludes(logicalServiceApi, 'describeLogicalServiceEnvironments', 'logical service environment API seam');
assertIncludes(logicalServiceApi, 'bindServiceEnvironment', 'explicit service environment binding seam');
assertIncludes(groupApi, 'describeConfigGroupEnvironments', 'config group environment API seam');
assertIncludes(fileApi, 'describeConfigFileEnvironments', 'config file environment API seam');
assertIncludes(fileApi, 'brief: true', 'config file environment query must return summaries');

assertIncludes(serviceDetail, 'EnvironmentResourceSwitcher', 'service detail environment switcher');
assertIncludes(serviceDetail, 'binding.service_name', 'service environment switching must carry runtime service name');
assertIncludes(serviceDetail, 'resolveServiceEnvironmentBinding', 'direct environment URL must resolve explicit binding');
assertIncludes(groupFiles, 'EnvironmentResourceSwitcher', 'config group Files environment switcher');
assertIncludes(fileView, 'EnvironmentResourceSwitcher', 'config file detail environment switcher');
assertIncludes(fileView, 'describeConfigFileEnvironments(currentGroup, currentName)', 'config file environment identity includes group and file name');

console.log('cross-environment resource views verified');
