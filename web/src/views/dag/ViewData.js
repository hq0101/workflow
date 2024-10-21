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
    children: [
      {
        id: 'dependency-checker',
        data: {
          label: 'Dependency Checker',
        },
      },
      {
        id: 'oss-license-checker',
        data: {
          label: 'OSS License Checker',
        },
      },
      {
        id: 'sca',
        data: {
          label: 'SCA',
        },
      },
    ],
  },
  {
    id: 'artefact-analysis',
    dependencies: ['static-analysis'],
    data: {
      label: 'Artefact Analysis',
    },
    children: [
      {
        id: 'image-hardening',
        data: {
          label: 'Image Hardening',
        },
      },
      {
        id: 'image-Scan',
        data: {
          label: 'Image Scan',
        },
      },
    ],
  },
  {
    id: 'deploy',
    dependencies: ['artefact-analysis'],
    data: {
      label: 'Deploy',
    },
    children: [
      {
        id: 'deployment',
        data: {
          label: 'Deployment',
        },
      },
    ],
  },
  {
    id: 'dynamic-analysis',
    dependencies: ['deploy'],
    data: {
      label: 'Dynamic Analysis',
    },
    children: [
      {
        id: 'dast',
        data: {
          label: 'DAST',
        },
      },
    ],
  },
  {
    id: 'end',
    dependencies: ['dynamic-analysis'],
  },
];
