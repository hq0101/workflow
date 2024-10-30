export const HierarchicalModel = [
  {
    id: 'start',
  },
  {
    id: 'git-clone',
    dependencies: ['start'],
    data: {
      label: 'Git Clone',
    },
  },
  {
    id: 'secrets-scanner',
    dependencies: ['git-clone'],
    data: {
      label: 'Secrets Scanner',
    },
  },
  {
    id: 'static-analysis',
    dependencies: ['secrets-scanner'],
    data: {
      label: 'Static Analysis',
    },
  },
  {
    id: 'dependency-checker',
    data: {
      label: 'Dependency Checker',
    },
    parentId: 'static-analysis'
  },
  {
    id: 'oss-license-checker',
    data: {
      label: 'OSS License Checker',
    },
    parentId: 'static-analysis'
  },
  {
    id: 'sca',
    data: {
      label: 'SCA',
    },
    parentId: 'static-analysis'
  },
  {
    id: 'artefact-analysis',
    dependencies: ['static-analysis'],
    data: {
      label: 'Artefact Analysis',
    },
  },
  {
    id: 'image-hardening',
    data: {
      label: 'Image Hardening',
    },
    parentId: 'artefact-analysis'
  },
  {
    id: 'image-Scan',
    data: {
      label: 'Image Scan',
    },
    parentId: 'artefact-analysis'
  },
  {
    id: 'deploy',
    dependencies: ['artefact-analysis'],
    data: {
      label: 'Deploy',
    },
  },
  {
    id: 'deployment',
    data: {
      label: 'Deployment',
    },
    parentId: 'deployment'
  },
  {
    id: 'dynamic-analysis',
    dependencies: ['deploy'],
    data: {
      label: 'Dynamic Analysis',
    },
  },
  {
    id: 'dast',
    data: {
      label: 'DAST',
    },
    parentId: 'dynamic-analysis'
  },
  {
    id: 'end',
    dependencies: ['dynamic-analysis'],
  },
];
