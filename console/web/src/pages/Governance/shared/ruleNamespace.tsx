import React from 'react';

const RuleNamespaceContext = React.createContext('');

export const RuleNamespaceProvider = RuleNamespaceContext.Provider;

export const useRuleNamespace = () => React.useContext(RuleNamespaceContext);
