export const Timeline = () => {
  return <div className="container mx-auto px-4 py-8">
    <div className="relative">
      <div className="absolute left-1.5 transform w-1 h-full bg-gray-200"></div>

      <div className="relative">
        <div className="flex items-center mb-8 flex-row" role="article">
          <div className="w-4 h-4 mr-2 rounded-2xl bg-gray-200"/>
          <div className="pr-2 text-left" role="button">
            <div className="p-4">
              <h3 className="text-lg font-bold text-gray-800">Отборочный этап</h3>
              <span className="text-xs text-md-dark-primary">Завершенный этап.</span>
            </div>
          </div>
        </div>
      </div>

      <div className="relative">
        <div className="flex items-center mb-8 flex-row" role="article">
          <div className="w-4 h-4 mr-2 rounded-2xl bg-gray-200" />
          <div className="pr-2 text-left" role="button">
            <div className="p-4">
              <h3 className="text-lg font-bold text-gray-800">Комната ожидания</h3>
              <span className="text-xs text-md-dark-primary">В процессе...</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
}
