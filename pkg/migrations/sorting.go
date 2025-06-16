package migrations

type DependencyGraph struct {
	actions    map[string]DBAction
	deps       dependencyMap
	dependants dependencyMap
}

type dependencyMap map[string]map[string]struct{}

func (g *DependencyGraph) AddActions(actions []DBAction, deps dependencyMap) {
	for _, action := range actions {
		g.actions[action.ID()] = action
	}

	for id, dep := range deps {
		for depID := range dep {
			g.deps[id][depID] = struct{}{}
		}
		for dependantID := range g.dependants {
			if _, ok := g.dependants[dependantID]; !ok {
				g.dependants[dependantID] = make(map[string]struct{})
			}
			g.dependants[dependantID][id] = struct{}{}
		}
	}
}

func (g *DependencyGraph) GetActions() []DBAction {
	var actions []DBAction
	for len(g.actions) > 0 {
		for _, action := range g.actions {
			if _, ok := g.deps[action.ID()]; !ok {
				actions = append(actions, action)
				delete(g.actions, action.ID())
			}
			if _, ok := g.dependants[action.ID()]; ok {
				for depID := range g.dependants[action.ID()] {
					delete(g.deps[depID], action.ID())
					if len(g.deps[depID]) == 0 {
						delete(g.deps, depID)
					}
				}
				delete(g.dependants, action.ID())
			}
		}
	}
	return actions
}
