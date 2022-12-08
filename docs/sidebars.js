/**
 * Creating a sidebar enables you to:
 - create an ordered group of docs
 - render a sidebar for each doc of that group
 - provide next/previous navigation

 The sidebars can be generated from the filesystem, or explicitly defined here.

 Create as many sidebars as you want.
 */

// @ts-check

/** @type {import('@docusaurus/plugin-content-docs').SidebarsConfig} */
const sidebars = {
  // By default, Docusaurus generates a sidebar from the docs folder structure
  // docsSidebar: [{ type: "autogenerated", dirName: "." }],

  docsSidebar: [
    {
      type: 'doc',
      id: 'README',
      label: 'Get started',
    },
    {
      type: 'doc',
      id: 'install',
      label: 'Install Rill',
    },
    {
      type: 'doc',
      id: 'import-data',
      label: 'Import data source',
    },
    {
      type: 'doc',
      id: 'sql-models',
      label: 'Model SQL transformations',
    },
    {
      type: 'doc',
      id: 'metrics-dashboard',
      label: 'Define metrics dashboard',
    },
    {
      type: 'doc',
      id: 'cli',
      label: 'CLI documentation',
    },
    {
      type: 'category',
      label: 'Contributors',
      items: ['contributors/development', 'contributors/local-testing', 'contributors/guidelines'],
    },
  ],
};

module.exports = sidebars;
